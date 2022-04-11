package config

import (
	"flag"
	"fmt"
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
	Token string `yaml:"token"`
}

type ScreenshotMonitor struct {
	Token string `yaml:"token"`
}

type Bitly struct {
	Token string `yaml:"token"`
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

func (c *Config) processError(err error) {
	log.Println(err)
	os.Exit(1)
}

func (c *Config) readYmlConfig() {
	var cfg Config

	args := ProcessArgs(&cfg)
	err := cleanenv.ReadConfig(args.ConfigPath, c)
	if err != nil {
		c.processError(err)
	}
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("parameters", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yml", "Path to configuration file")

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	f.Parse(os.Args[1:])
	return a
}

func init() {
	conf.readYmlConfig()
}

func GetConfig() Config {
	return conf
}
