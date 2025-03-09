package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"sync"
)

var (
	once   sync.Once
	config *Config
)

type Config struct {
	Port     int `required:"true" default:"3000"`
	Swagger  Swagger
	Channels string `required:"true"`
	Mongo    Mongo
}

type Swagger struct {
	HostName string `required:"true" split_words:"true" default:"localhost"`
}

type Mongo struct {
	Url string `split_words:"true" required:"true" default:"mongodb://user:123@localhost:27017"`
	DB  string `split_words:"true" required:"true" default:"db"`
}

func Get() Config {
	once.Do(func() {
		config = &Config{}
		if err := envconfig.Process("", config); err != nil {
			panic(fmt.Sprintf("Error loading config: %#v", err))
		}
	})
	return *config
}
