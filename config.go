package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type App struct {
	Mode    string `yaml:"mode"`
	LogFile string `yaml:"log-file"`
}
type Http struct {
	Port string `yaml:"port"`
	SSL  struct {
		Crt string `yaml:"crt"`
		Key string `yaml:"key"`
	} `yaml:"ssl"`
}

type Config struct {
	App  App  `yaml:"app"`
	Http Http `yaml:"http"`
}

var config = &Config{}

func init() {
	var conf string
	if len(os.Args) == 2 {
		conf = os.Args[1]
	}
	if conf == "" {
		conf = "conf.yaml"
	}

	data, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Println("Config file not found, use default configs.")
		config = &Config{
			App: App{
				Mode:    "debug",
				LogFile: "oktools.log",
			},
			Http: Http{
				Port: "9175",
			},
		}
	}

	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		log.Fatalln(err)
	}
}
