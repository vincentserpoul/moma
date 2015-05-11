package utils

import (
	"encoding/json"
	"io/ioutil"
)

// RedisConfig contains connection params
type RedisConfig struct {
	Host string
	Port string
}

type ApplicationConfig struct {
	Redis      RedisConfig
	Port       string
	PersonaURL string
}

// LoadConfig Config loader from json file
func LoadConfig(fileName string) (ApplicationConfig, error) {
	var config ApplicationConfig

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ApplicationConfig{}, err
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return ApplicationConfig{}, err
	}

	return config, nil

}
