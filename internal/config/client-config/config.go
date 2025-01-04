package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ConfigEnv         = "CONFIG_PATH"
	DefaultConfigPath = "./config/client-config.yaml"
)

type Config struct {
	FetcherConfig `yaml:"fetcher" json:"fetcher"`
	AppConfig     `yaml:"client" json:"client"`
}

type FetcherConfig struct {
	UpdateTime int64 `yaml:"update_time" json:"update_time"`
}

type AppConfig struct {
	ReportInterval int64        `yaml:"report_interval" json:"report_interval"`
	PollingRate    int64        `yaml:"polling_rate" json:"polling_rate"`
	ReportURL      string       `yaml:"report_url" json:"report_url"`
	ClientConfig   ClientConfig `yaml:"server" json:"server"`
	SaverConfig    `yaml:"saver" json:"saver"`
}

// TODO: переделать это говно
type SaverConfig struct {
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	Url     string        `yaml:"url" json:"url"`
}

type ClientConfig struct {
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

func New() *Config {
	const fn = "cfg.New"

	configPath := os.Getenv(ConfigEnv)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file %s not found", configPath))
	}

	switch ext := filepath.Ext(configPath); ext {
	case ".json":
		return parseConfigFromJSON(configPath)
	case ".yaml":
		return parseConfigFromYAML(configPath)
	default:
		panic(fmt.Sprintf("Not supported config file %s: %s", fn, ext))
	}
}

func parseConfigFromJSON(configPath string) *Config {
	const fn = "cfg.parseConfigFromJSON"

	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("%v: %v", fn, err))
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		panic(fmt.Sprintf("%v: %v", fn, err))
	}

	return &config
}

func parseConfigFromYAML(configPath string) *Config {
	const fn = "cfg.parseConfigFromYAML"

	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("%v: %v", fn, err))
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		panic(fmt.Sprintf("%v: %v", fn, err))
	}

	return &config
}
