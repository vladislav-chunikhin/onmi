package config

import (
	"time"
)

type Config struct {
	ClientsConfig ClientsConfig `yaml:"clients"`
}

type ClientsConfig struct {
	SupperApp ClientConfig `yaml:"supperApp"`
}

type ClientConfig struct {
	Host    string        `yaml:"host"`
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
	Delta   time.Duration `yaml:"delta"`
}
