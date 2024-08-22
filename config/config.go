package config

import "gopkg.in/yaml.v3"

type Instrument struct {
	FIGI    string `yaml:"figi"`
	Name    string `yaml:"name"`
	Channel string `yaml:"channel"`
}

type Config struct {
	Token       string       `yaml:"token" validate:"required"`
	Instruments []Instrument `yaml:"instruments" validate:"required"`
	RedisURL    string       `yaml:"redis_url" validate:"required"`
	Log         Log          `yaml:"log" validate:"required"`
}

func (c *Config) Dump() ([]byte, error) {
	return yaml.Marshal(c)
}
