package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	internalhttp "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/http"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	HTTPServer internalhttp.Conf `toml:"http"`
	Storage    StorageConf       `toml:"storage"`
	Logger     LoggerConf        `toml:"logger"`
}

type LoggerConf struct {
	Level string `toml:"level"`
}

type StorageConf struct {
	DB  string `toml:"db"`
	DSN string `toml:"dsn"`
}

type HTTPServerConf struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) LoadConfigFile(filename string) error {
	filedata, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read config file: %s", err.Error())
	}

	err = toml.Unmarshal(filedata, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}
