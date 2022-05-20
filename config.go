package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	configFileLocations = []string{
		"~/.config/simplesonic/simplesonic.json",
		"/etc/simplesonic/simplesonic.json",
	}
	configDefaultValues = &SimplesonicConfig{
		Server: &ServerConfig{
			ListenAddress: ":4040",
		},
	}
	Config = configDefaultValues.readConfigFile()
)

type SimplesonicConfig struct {
	Server         *ServerConfig        `json:"server"`
	MusicFolders   []*MusicFolderConfig `json:"musicFolders"`
	PlaylistFolder string               `json:"playlistFolder"`
	Users          []*UserConfig        `json:"users"`
	MPD            *MPDConfig           `json:"mpd"`
}

type ServerConfig struct {
	ListenAddress string `json:"listenAddress"`
	TLSKey        string `json:"tlsKey"`
	TLSCert       string `json:"tlsCert"`
}

type MusicFolderConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type UserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MPDConfig struct {
	UnixSocket string `json:"unixSocket"`
}

func (config *SimplesonicConfig) readConfigFile() *SimplesonicConfig {
	var configFile string
	for _, configFileLocation := range configFileLocations {
		if IsExists(configFileLocation) {
			configFile = configFileLocation
		}
	}
	if configFile == "" {
		ProcessErrorArg(fmt.Fprintf(os.Stderr, "A simplesonic config file has to exist "+
			"in one of the following locations:\n - %s\n", strings.Join(configFileLocations, "\n - ")))
		os.Exit(1)
	}
	file := ProcessErrorArg(os.Open(configFile)).(*os.File)
	defer Close(file)
	ProcessError(json.NewDecoder(file).Decode(&config))
	if config.Server.TLSKey != "" && !IsExists(config.Server.TLSKey) {
		if tlsKey := filepath.Join(filepath.Dir(configFile), config.Server.TLSKey); IsExists(tlsKey) {
			config.Server.TLSKey = tlsKey
		} else {
			ProcessErrorArg(fmt.Fprintf(os.Stderr, "TLS key does not exist: %s\n", config.Server.TLSKey))
			os.Exit(1)
		}
	}
	if config.Server.TLSCert != "" && !IsExists(config.Server.TLSCert) {
		if tlsCert := filepath.Join(filepath.Dir(configFile), config.Server.TLSCert); IsExists(tlsCert) {
			config.Server.TLSCert = tlsCert
		} else {
			ProcessErrorArg(fmt.Fprintf(os.Stderr, "TLS cert does not exist: %s\n", config.Server.TLSCert))
			os.Exit(1)
		}
	}
	return config
}
