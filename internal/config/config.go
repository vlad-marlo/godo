package config

import (
	"log"
	"sync"

	"github.com/caarlos0/env/v6"
)

type HTTPConfig struct {
	BindAddr string `env:"HTTP_SERVER_ADDRESS"`
	Port     int    `env:"HTTP_PORT"`
}

type Config struct {
	HTTP HTTPConfig
}

var c *Config
var once sync.Once

func Get() *Config {
	once.Do(func() {
		c = &Config{}
		if err := env.Parse(c); err != nil {
			log.Fatalf("got unexpected error while getting new config instance: %v", err)
		}
	})
	return c
}

func (c *Config) Copy() *Config {
	return &Config{
		HTTP: HTTPConfig{
			BindAddr: c.HTTP.BindAddr,
			Port:     c.HTTP.Port,
		},
	}
}
