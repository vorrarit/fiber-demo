package main

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Application struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
		Otel struct {
			Enable   bool   `yaml:"enable" envconfig:"OTEL_ENABLE"`
			Grpc_Url string `yaml:"grpc_url" envconfig:"OTEL_GRPC_URL"`
		} `yaml:"otel"`
	} `yaml:"application"`
	ServiceB struct {
		Url string `yaml:"url"`
	} `yaml:"service_b"`
}

func ReadConfig() Config {
	var config Config
	err := readFile(&config)
	if err != nil {
		fmt.Println("Error reading config config.yaml - %+v\n", err)
		os.Exit(2)
	}
	err = readEnv(&config)
	if err != nil {
		fmt.Println("Error reading config from environment - %+v\n", err)
		os.Exit(2)
	}
	return config
}

func readFile(config *Config) error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}

func readEnv(config *Config) error {
	err := envconfig.Process("", config)
	if err != nil {
		return err
	}
	return nil
}
