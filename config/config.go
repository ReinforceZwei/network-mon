package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var DefaultFileName = "netmonconfig.json"

type AppConfig struct {
	Router                string
	SshUser               string
	SshPassword           string
	TestTarget            []string
	TestIntervalSecond    int
	FailingIntervalSecond int
	MaxFailAttempt        int
	NetworkRestartCommand string
	RestartAttempt        int
	RestartMaxWaitTime    int
	RouterRebootCommand   string
	RebootAttempt         int
	RebootMaxWaitTime     int
	Notify                DiscordWebhookConfig
}

type DiscordWebhookConfig struct {
	Url            string
	MessagePayload string
}

func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Router:      "192.168.1.1",
		SshUser:     "admin",
		SshPassword: "admin",
		TestTarget: []string{
			"1.1.1.1",
			"8.8.8.8",
			"8.8.4.4",
		},
		TestIntervalSecond:    300,
		FailingIntervalSecond: 10,
		MaxFailAttempt:        10,
		NetworkRestartCommand: "service restart_wan",
		RestartAttempt:        3,
		RestartMaxWaitTime:    40,
		RouterRebootCommand:   "reboot",
		RebootAttempt:         1,
		RebootMaxWaitTime:     120,
		Notify: DiscordWebhookConfig{
			Url:            "",
			MessagePayload: "{\"content\":\"%s\"}",
		},
	}
}

func (c *AppConfig) Save() error {
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(DefaultFileName, b, 0666)
	if err != nil {
		return err
	}
	return nil
}

func Load() (*AppConfig, error) {
	b, err := ioutil.ReadFile(DefaultFileName)
	if err != nil {
		return nil, err
	}
	// Must be initalized
	c := &AppConfig{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func LoadOrCreateDefault() (*AppConfig, error) {
	var c *AppConfig
	if _, err := os.Stat(DefaultFileName); errors.Is(err, os.ErrNotExist) {
		// Create new config file
		c = NewDefaultConfig()
		c.Save()
		fmt.Println("A default config file was created. Please review and edit the config.\nExiting...")
		os.Exit(1)
	} else {
		// Load config file
		c, err = Load()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
