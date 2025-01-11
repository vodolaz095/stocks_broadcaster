package config

import (
	"github.com/vodolaz095/pkg/zerologger"
	"gopkg.in/yaml.v3"
)

// Input defines reader's configuration which helps it to obtain last price data
type Input struct {
	// Name makes this input different
	Name string `yaml:"name"  validate:"required"`
	// Token is required to access OpenInvestement API - see readme how to obtain it
	Token string `yaml:"token" validate:"required"`
	// Figis defines list of instruments' last price feeds this reader will subscribe on
	Figis []string `yaml:"figis" validate:"required"`
	// LocalAddr makes reader use this local ip address to dial InvestAPI - so, if your machine
	// has few network interfaces you can use any of them with seprate Token and separate list of
	// instruments Figis to be observed
	LocalAddr string `yaml:"local_addr" validate:"omitempty,cidr"`
}

// Output defines redis servers configuration where we broadcast last price data from Input
type Output struct {
	Name     string `yaml:"name" validate:"required"`
	RedisURL string `yaml:"redis_url" validate:"required"`
}

// Instrument defines routing parameters (where we broadcast messages) and
// how to provide instrument name in last price update
type Instrument struct {
	FIGI    string `yaml:"figi" validate:"required"`
	Name    string `yaml:"name" validate:"required"`
	Channel string `yaml:"channel" validate:"required"`
}

// Config defines structure we expect in configuration file of application
type Config struct {
	Inputs      []Input        `yaml:"inputs" validate:"required"`
	Instruments []Instrument   `yaml:"instruments" validate:"required"`
	Outputs     []Output       `yaml:"outputs" validate:"required"`
	Log         zerologger.Log `yaml:"log" validate:"required"`
}

// Dump writes current runtime config
func (c *Config) Dump() ([]byte, error) {
	return yaml.Marshal(c)
}
