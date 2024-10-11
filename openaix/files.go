package openaix

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

type Files struct {
	tele     *telebot.Bot
	logger   *logrus.Logger
	filePath string
}

func NewFiles(tele *telebot.Bot, logger *logrus.Logger) Files {
	if tele == nil {
		return Files{}
	}
	return Files{
		tele:   tele,
		logger: logger,
	}
}

func (f *Files) DownloadAsync(ctx context.Context, file *telebot.File, extension string) {
	var wg sync.WaitGroup
	select {
	case <-ctx.Done():
		f.logger.Info("shut down")
		return
	default:
		wg.Add(1)
		go func(file *telebot.File, extenstion string) {
			defer wg.Done()
			f.download(file, extenstion)
		}(file, extension)
	}
	wg.Wait()
	f.logger.Info("done")
}

func (f *Files) download(file *telebot.File, extension string) {
	reader, err := f.tele.File(file)
	if err != nil {
		return
	}
	defer reader.Close()

	root, err := os.Getwd()
	if err != nil {
		return
	}
	fileDir := filepath.Join(root, "files")
	if err := os.MkdirAll(fileDir, os.ModePerm); err != nil {
		return
	}

	filePath := filepath.Join(fileDir, fmt.Sprintf("%s.%s", file.FileID, extension))
	out, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		return
	}
	f.filePath = filePath
}

func (f *Files) Filepath() string {
	return f.filePath
}

func (f *Files) Cleanup() {
	root := filepath.Join(".", "files")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			f.logger.Infof("directory %s not found", root)
			return
		}
		f.logger.Errorf("failed to read directory %s: %s", root, err)
		return
	}

	for _, entry := range entries {
		filePath := filepath.Join(root, entry.Name())
		if err := os.Remove(filePath); err != nil {
			f.logger.Errorf("failed to remove file %s: %s", filePath, err)
		}
	}
	f.logger.Infof("removed %d files from %s", len(entries), root)
}
