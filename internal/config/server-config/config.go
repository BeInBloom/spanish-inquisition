package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	ErrUnexpectedFlag = errors.New("unexpected flag")
)

const (
	ConfigEnv         = "CONFIG_PATH"
	DefaultConfigPath = "./config/config.yaml"
)

type Config struct {
	ServerConfig `yaml:"server" json:"server"`
	EnvConfig    `yaml:"env" json:"env"`
	DBConfig     `yaml:"database" json:"database"`
}

type DBConfig struct {
	Address   string `yaml:"address" json:"address" env:"DATABASE_DSN"`
	BakConfig `yaml:"bakconfig" json:"bakconfig"`
}

type ServerConfig struct {
	Address string `yaml:"address" json:"address" env:"ADDRESS"`
	// Port        int           `yaml:"port" json:"port"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout" env:"TIMEOUT"`
	IdleTimeout time.Duration `yaml:"idle_timeout" json:"idle_timeout" env:"IDLE_TIMEOUT"`
	Restore     bool          `yaml:"restore" json:"restore" env:"RESTORE"`
}

type EnvConfig struct {
	Env string `yaml:"env" json:"env"`
}

type BakConfig struct {
	Path          string `yaml:"path" json:"path" env:"FILE_STORAGE_PATH"`
	StoreInterval int    `yaml:"store_interval" json:"store_interval" env:"STORE_INTERVAL"`
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

	pflag.StringVarP(&config.ServerConfig.Address, "address", "a", "localhost:8080", "server address")
	pflag.DurationVarP(&config.ServerConfig.Timeout, "timeout", "t", 10*time.Second, "server request timeout")
	pflag.DurationVarP(&config.ServerConfig.IdleTimeout, "idle-timeout", "i", 10*time.Second, "server idle timeout")
	pflag.BoolVarP(&config.ServerConfig.Restore, "restore", "r", true, "restore database")

	pflag.StringVarP(&config.EnvConfig.Env, "env", "e", "dev", "environment")

	pflag.StringVarP(&config.BakConfig.Path, "db-path", "d", "./pesiks_better_then_kitiks.txt", "database path")
	pflag.IntVarP(&config.BakConfig.StoreInterval, "store-interval", "s", 300, "store interval")
	pflag.StringVarP(&config.DBConfig.Address, "db-address", "d", "", "database address")

	pflag.Parse()

	if pflag.NArg() > 0 {
		fmt.Printf("%s", ErrUnexpectedFlag)
		os.Exit(1)
	}

	return &config
}

func checkEnvServerConfig(config *ServerConfig) {
	var envConfig ServerConfig

	if err := env.Parse(&envConfig); err != nil {
		return
	}

	if envConfig.Address != "" {
		config.Address = envConfig.Address
	}

	if envConfig.Timeout != 0 {
		config.Timeout = envConfig.Timeout
	}

	if envConfig.IdleTimeout != 0 {
		config.IdleTimeout = envConfig.IdleTimeout
	}

	if envConfig.Restore {
		config.Restore = envConfig.Restore
	}
}

func checkEnvBakConfig(config *BakConfig) {
	var envConfig BakConfig

	if err := env.Parse(&envConfig); err != nil {
		return
	}

	if envConfig.Path != "" {
		config.Path = envConfig.Path
	}

	//TODO потенциальная проблема
	if envConfig.StoreInterval != 0 {
		config.StoreInterval = envConfig.StoreInterval
	}
}

func checkEnvDatabaseConfig(config *DBConfig) {
	var envConfig DBConfig

	if err := env.Parse(&envConfig); err != nil {
		return
	}

	if envConfig.Address != "" {
		config.Address = envConfig.Address
	}
}

func getEnvAndFlagConfig() *Config {
	config := parseConfigFromFlags()
	checkEnvServerConfig(&config.ServerConfig)
	checkEnvBakConfig(&config.BakConfig)
	checkEnvDatabaseConfig(&config.DBConfig)

	return config
}
