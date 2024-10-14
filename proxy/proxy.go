package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

var (
	errEmptyFilepath         = errors.New("filepath is empty")
	errFailedToReadFile      = errors.New("failed to read file")
	errFailedToUnmarshalFile = errors.New("failed to unmarshal file")
)

// example of parsed proxy obj below
// https://username:password@192.168.1.1:8080
type Proxy struct {
	Type string `json:"type"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

func (p Proxy) FromFile(filepath string) ([]Proxy, error) {
	if filepath == "" {
		return nil, errEmptyFilepath
	}
	jsnFile, err := os.ReadFile(path.Join(".", "proxy.json"))
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		return nil, errFailedToReadFile
	}
	var wrapper struct {
		Proxies []Proxy `json:"proxies"`
	}
	if err := json.Unmarshal(jsnFile, &wrapper); err != nil {
		fmt.Printf("error unmarshaling file: %v\n", err)
		return nil, errFailedToUnmarshalFile
	}
	return wrapper.Proxies, nil
}

func (p Proxy) String() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d", p.Type, p.User, p.Pass, p.Ip, p.Port)
}
