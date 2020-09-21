package domain

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Urls []URL `json:"urls"`
}

type Topic struct {
	Topic string `json:"topic"`
}

type HTTP struct {
	Host string `json:"host"`
}

type URL struct {
	Method string `json:"method"`
	Nsq    *Topic `json:"nsq"`
	HTTP   *HTTP  `json:"http"`
	Path   string `json:"path"`
}

func ParseFileConfig(filename string) (Config, error) {
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return Config{}, err
	}

	c := Config{}

	err = yaml.Unmarshal(source, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
