package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	ErrUnexpectedFlag = errors.New("unexpected flag")
)

const (
	ConfigEnv         = "CONFIG_PATH"
	DefaultConfigPath = "./config/client-config.yaml"
)

type Config struct {
	SaverConfig  SaverConfig `yaml:"saver" json:"saver"`
	PollInterval int         `yaml:"polling" json:"polling" env:"POLL_INTERVAL"`
	AppConfig    AppConfig   `yaml:"app" json:"app"`
}

// // TODO: переделать это говно
type SaverConfig struct {
	Timeout int    `yaml:"timeout" json:"timeout" env:"SAVER_TIMEOUT"`
	URL     string `yaml:"url" json:"url" env:"ADDRESS"`
	Key     string `yaml:"key" json:"key" env:"KEY"`
}

type AppConfig struct {
	ReportInterval int `yaml:"report_interval" json:"report_interval" env:"REPORT_INTERVAL"`
}

func New() *Config {
	const fn = "cfg.New"

	configPath := os.Getenv(ConfigEnv)
	if configPath == "" {
		return getEnvAndFlagConfig()
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

func parseConfigFromFlags() *Config {
	var config Config

	pflag.StringVarP(&config.SaverConfig.URL, "address", "a", "localhost:8080", "server address")
	pflag.IntVarP(&config.AppConfig.ReportInterval, "report-interval", "r", 10, "report interval")
	pflag.IntVarP(&config.PollInterval, "poll-interval", "p", 10, "polling interval")
	pflag.IntVarP(&config.SaverConfig.Timeout, "timeout", "t", 10, "timeout")
	pflag.Parse()

	if pflag.NArg() > 0 {
		fmt.Printf("%s", ErrUnexpectedFlag)
		os.Exit(1)
	}

	return &config
}

func getEnvAndFlagConfig() *Config {
	config := parseConfigFromFlags()
	checkEnvConfig(config)
	return config
}

func checkEnvConfig(config *Config) {
	var envConfig Config

	if err := env.Parse(&envConfig); err != nil {
		return
	}

	if envConfig.PollInterval != 0 {
		config.PollInterval = envConfig.PollInterval
	}

	if envConfig.AppConfig.ReportInterval != 0 {
		config.AppConfig.ReportInterval = envConfig.AppConfig.ReportInterval
	}

	if envConfig.SaverConfig.Timeout != 0 {
		config.SaverConfig.Timeout = envConfig.SaverConfig.Timeout
	}

	if envConfig.SaverConfig.URL != "" {
		config.SaverConfig.URL = envConfig.SaverConfig.URL
	}
}
