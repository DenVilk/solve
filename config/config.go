package config

import (
	"encoding/json"
	"io/ioutil"
)

// Configuration for Solve API and Invoker
type Config struct {
	Server   Server   `json:""`
	DB       DB       `json:""`
	Security Security `json:""`
}

// Loads configuration from json file
func LoadFromFile(file string) (cfg Config, err error) {
	bytes, err := ioutil.ReadFile(file)
	if err == nil {
		err = json.Unmarshal(bytes, &cfg)
	}
	return
}
