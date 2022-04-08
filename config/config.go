package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

var conf Config

type Config struct {
	Planfix           Planfix           `yaml:"planfix"`
	Dropbox           Dropbox           `yaml:"dropbox"`
	ScreenshotMonitor ScreenshotMonitor `yaml:"sm"`
	Bitly             Bitly             `yaml:"bitly"`
	Users             map[int]int       `yaml:"users"`
}

type Planfix struct {
	Account    string `yaml:"account"`
	ApiKey     string `yaml:"api_key"`
	ApiSecret  string `yaml:"api_secret"`
	User       string `yaml:"user"`        // user that posts updates to planfix
	Password   string `yaml:"password"`    // given user password
	AnalyticId int    `yaml:"analytic_id"` //
}

type Dropbox struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Token        string `yaml:"token"`
}

type ScreenshotMonitor struct {
	Token string `yaml:"token"`
}

type Bitly struct {
	Token string `yaml:"token"`
}

func (c *Config) processError(err error) {
	log.Println(err)
	os.Exit(1)
}

func (c *Config) readYmlConfig() {
	err := cleanenv.ReadConfig("config/config.yml", c)
	if err != nil {
		c.processError(err)
	}
}

func init() {
	conf.readYmlConfig()
}

func GetConfig() Config {
	return conf
}
