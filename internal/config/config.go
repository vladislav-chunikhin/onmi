package config

import (
	"time"

	"onmi/pkg/circuitbreaker"
)

type Config struct {
	ClientsConfig ClientsConfig `yaml:"clients"`
}

type ClientsConfig struct {
	SupperApp ClientConfig `yaml:"supperApp"`
}

type ClientConfig struct {
	Host           string                `yaml:"host"`
	Port           string                `yaml:"port"`
	Timeout        time.Duration         `yaml:"timeout"`
	CircuitBreaker circuitbreaker.Config `yaml:"circuitBreaker"`
}
