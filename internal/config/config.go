package config

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

const (
	appName = "go-base-app"
)

type BaseConfig struct {
	LogLevel       string `json:"LogLevel"`
	PprofPort      string `json:"PprofPort"`
	GrpcPort       int    `json:"GrpcPort"`
	APIPort        int    `json:"ApiPort"`
	PromethiusPort int    `json:"PromethiusPort"`
}

func GetConfiguration() *BaseConfig {
	var config BaseConfig
	err := defaultViper.Unmarshal(&config)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to unmarshal configuration: %s", err))
	}
	return &config
}

func getDefaultConfiguration() BaseConfig {
	return BaseConfig{
		LogLevel:       "warning",
		PprofPort:      ":10001",
		GrpcPort:       10000,
		APIPort:        11000,
		PromethiusPort: 12000,
	}
}

var defaultViper = viper.New()

func InitConfiguration(path string) error {
	defaultViper.SetConfigName("config")
	defaultViper.SetConfigType("yml")

	replacer := strings.NewReplacer(".", "_")
	defaultViper.SetEnvKeyReplacer(replacer)

	dfltConfig := getDefaultConfiguration()
	yamlConfig, err := yaml.Marshal(dfltConfig)
	if err != nil {
		fmt.Println("Error parsing default configuration")
		return err
	}

	yamlReader := bytes.NewReader(yamlConfig)
	if err := defaultViper.ReadConfig(yamlReader); err != nil {
		fmt.Println("Error reading default configuration")
		return err
	}

	// Is there a config file available?
	// - viper will search paths in the order they are added
	// - last found wins

	defaultViper.AddConfigPath(filepath.Join("/", "etc", appName))
	if path != "" {
		defaultViper.AddConfigPath(path)
	}
	defaultViper.AutomaticEnv()

	if err := defaultViper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("No configuration file found - using defaults %s\n", err.Error())
		} else {
			fmt.Printf("Error reading config file: %s\n", err.Error())
			return err
		}
	}

	//Have it watch the config file for changes
	viper.OnConfigChange(OnConfigChangeEvent)
	defaultViper.WatchConfig()

	return nil
}

func OnConfigChangeEvent(event fsnotify.Event) {
	fmt.Println("Config Changed")
}

func GetRpcPort() int {
	return GetConfiguration().GrpcPort
}

func GetPrometheusPort() int {
	return GetConfiguration().PromethiusPort
}

func GetRpcGatewayPort() int {
	return GetConfiguration().APIPort
}
