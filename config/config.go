package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/tibeahx/gpt-helper/logger"
	"gopkg.in/yaml.v3"
)

var cfg *Config
var log = logger.GetLogger()

type Config struct {
	BotApikey     string
	AiApikey      string
	RotationDelay string `yaml:"rotationDelay"`
	BaseUrl       string `yaml:"baseUrl"`
}

func readConfig(filepath string) *Config {
	cfgFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Errorf("failed to read cfgfile: %v", err)
		return nil
	}
	var config Config
	if err := yaml.Unmarshal(cfgFile, &config); err != nil {
		log.Errorf("failed to unmarshal cfgfile: %v", err)
		return nil
	}
	return &config
}

func LoadConfig(filepath string) *Config {
	if filepath != "" && cfg == nil {
		cfg = readConfig(filepath)
	}
	if err := godotenv.Load(); err != nil {
		log.Error(err)
		return nil
	}
	cfg.AiApikey = os.Getenv("AI_TOKEN")
	cfg.BotApikey = os.Getenv("BOT_TOKEN")
	return cfg
}
