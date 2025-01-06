package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	ConfigEnv         = "CONFIG_PATH"
	DefaultConfigPath = "./config/client-config.yaml"
)

type Config struct {
	SaverConfig  SaverConfig `yaml:"saver" json:"saver"`
	PollInterval int         `yaml:"polling" json:"polling"`
	AppConfig    AppConfig   `yaml:"app" json:"app"`
}

// type FetcherConfig struct {
// 	UpdateTime int64 `yaml:"update_time" json:"update_time"`
// }

// type AppConfig struct {
// 	ReportInterval int64        `yaml:"report_interval" json:"report_interval"`
// 	PollingRate    int64        `yaml:"polling_rate" json:"polling_rate"`
// 	ReportURL      string       `yaml:"report_url" json:"report_url"`
// 	ClientConfig   ClientConfig `yaml:"server" json:"server"`
// 	SaverConfig    `yaml:"saver" json:"saver"`
// }

// // TODO: переделать это говно
type SaverConfig struct {
	Timeout int    `yaml:"timeout" json:"timeout"`
	URL     string `yaml:"url" json:"url"`
}

type AppConfig struct {
	ReportInterval int `yaml:"report_interval" json:"report_interval"`
}

// type ClientConfig struct {
// 	Timeout time.Duration `yaml:"timeout" json:"timeout"`
// }

func New() *Config {
	const fn = "cfg.New"

	configPath := os.Getenv(ConfigEnv)
	if configPath == "" {
		return parseConfigFromFlags()
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

	return &config
}
