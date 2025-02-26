package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

var conf Config

type Config struct {
	Planfix           Planfix           `yaml:"planfix"`
	Dropbox           Dropbox           `yaml:"dropbox"`
	ScreenshotMonitor ScreenshotMonitor `yaml:"sm"`
	Bitly             Bitly             `yaml:"bitly"`
	Links             Links             `yaml:"links"`
	Users             map[int]int       `yaml:"users"`
	Args              Args              `yaml:"-"`
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
	Token string `yaml:"token"`
}

type ScreenshotMonitor struct {
	Token string `yaml:"token"`
}

type Bitly struct {
	Token string `yaml:"token"`
}

type Links struct {
	Code   string `yaml:"code"`
	ApiUrl string `yaml:"api_url"`
}

// Args command-line parameters
type Args struct {
	ConfigPath string
	StartDate  string
	EndDate    string
}

func (c *Config) processError(err error) {
	log.Println(err)
	os.Exit(1)
}

func (c *Config) readYmlConfig() {
	flag.StringVar(&c.Args.ConfigPath, "config", "config.yml", "Path to config file")
	flag.StringVar(&c.Args.StartDate, "start", "", "Date to start from (YYYY-MM-DD). Defaults to yesterday.")
	flag.StringVar(&c.Args.EndDate, "end", "", "End date (YYYY-MM-DD) non-inclusive. Defaults to today.")
	flag.Parse()

	log.Println("Using config file", c.Args.ConfigPath)
	err := cleanenv.ReadConfig(c.Args.ConfigPath, c)
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
