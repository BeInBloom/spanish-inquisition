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
	DefaultConfigPath = "./config/config.yaml"
)

type Config struct {
	ServerConfig   `yaml:"server" json:"server"`
	EnvConfig      `yaml:"env" json:"env"`
	DatabaseConfig `yaml:"database" json:"database"`
}

type ServerConfig struct {
	Address     string        `yaml:"address" json:"address"`
	Port        int           `yaml:"port" json:"port"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

type EnvConfig struct {
	Env string `yaml:"env" json:"env"`
}

type DatabaseConfig struct{}

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
