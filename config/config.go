package config

import "gopkg.in/yaml.v3"

type Input struct {
	Name  string   `yaml:"name"  validate:"required"`
	Token string   `yaml:"token" validate:"required"`
	Figis []string `yaml:"figis" validate:"required"`
}

type Output struct {
	Name     string `yaml:"name" validate:"required"`
	RedisURL string `yaml:"redis_url" validate:"required"`
}

type Instrument struct {
	FIGI    string `yaml:"figi" validate:"required"`
	Name    string `yaml:"name" validate:"required"`
	Channel string `yaml:"channel" validate:"required"`
}

type Config struct {
	Inputs      []Input      `yaml:"inputs" validate:"required"`
	Instruments []Instrument `yaml:"instruments" validate:"required"`
	Outputs     []Output     `yaml:"outputs" validate:"required"`
	Log         Log          `yaml:"log" validate:"required"`
}

func (c *Config) Dump() ([]byte, error) {
	return yaml.Marshal(c)
}
