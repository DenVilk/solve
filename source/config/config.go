package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Server   ServerConfig   `json:""`
	Database DatabaseConfig `json:""`
	Security SecurityConfig `json:""`
}

// Loads configuration from json file
func LoadFromFile(file string) (cfg Config, err error) {
	bytes, err := ioutil.ReadFile(file)
	if err == nil {
		err = json.Unmarshal(bytes, &cfg)
	}
	return
}
