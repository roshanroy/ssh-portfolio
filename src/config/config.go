package config

import (
	"os"
	"log"
	"gopkg.in/yaml.v3"
	)

var GCfg GlobalConfig

type GlobalConfig struct {

	App struct {
		Port string `yaml:"port"`
		PrivateSSHKeyFile string `yaml:"privateSSHKeyFile"`
	} `yaml:"app"`
}

func init(){
	fileContentBytes,err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Unable to read config.yaml: ",err)	
	}

	err = yaml.Unmarshal(fileContentBytes,&GCfg)
	if err != nil {
                log.Fatal("Error reading config.yaml: ",err)
        }
}
