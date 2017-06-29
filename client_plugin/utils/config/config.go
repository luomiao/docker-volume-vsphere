// Copyright 2016-2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

// Read the plugin configuration file. The file is stored in JSON.
// See default-config.json at the root of the project.

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/natefinch/lumberjack"
	"github.com/vmware/docker-volume-vsphere/client_plugin/utils/log_formatter"
	"io/ioutil"
	"os"
	"runtime"
)

const (
	// VMDKPlugin is the name for vmdk plugin
	VMDKPlugin = "vmdk-plugin"
	// SharedPlugin is the name for shared plugin
	SharedPlugin = "shared-plugin"

	// PhotonDriver is a vmdk plugin driver for photon platform
	PhotonDriver = "photon"
	// VMDKDriver is a vmdk plugin driver for vSphere platform (deprecated)
	VMDKDriver = "vmdk"
	// VSphereDriver is a vmdk plugin driver for vSphere platform
	VSphereDriver = "vsphere"

	// SharedDriver is a shared plugin driver
	SharedDriver = "shared"

	// defaultPort is the default ESX service port.
	defaultPort = 1019

	// Local constants
	defaultMaxLogSizeMb  = 100
	defaultMaxLogAgeDays = 28
	defaultLogLevel      = "info"
)

// Config stores the configuration for the plugin
type Config struct {
	Driver        string `json:",omitempty"`
	LogPath       string `json:",omitempty"`
	MaxLogSizeMb  int    `json:",omitempty"`
	MaxLogAgeDays int    `json:",omitempty"`
	LogLevel      string `json:",omitempty"`
	Target        string `json:",omitempty"`
	Project       string `json:",omitempty"`
	Host          string `json:",omitempty"`
	Port          int    `json:",omitempty"`
	UseMockEsx    bool   `json:",omitempty"`
}

// Load the configuration from a file and return a Config.
func load(path string) (Config, error) {
	jsonBlob, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := json.Unmarshal(jsonBlob, &config); err != nil {
		return Config{}, err
	}
	setDefaults(&config)
	return config, nil
}

// setDefaults for any config setting that is at its `bottom`
func setDefaults(config *Config) {
	if config.MaxLogSizeMb == 0 {
		config.MaxLogSizeMb = defaultMaxLogSizeMb
	}
	if config.MaxLogAgeDays == 0 {
		config.MaxLogAgeDays = defaultMaxLogAgeDays
	}
	if config.LogLevel == "" {
		config.LogLevel = defaultLogLevel
	}
}

// LogInit init log with passed logLevel (and get config from configFile if it's present)
// returns True if using defaults,  False if using config file
func LogInit(logLevel *string, logFile *string, defaultLogFile string, configFile *string) bool {
	usingConfigDefaults := false
	c, err := load(*configFile)
	if err != nil {
		if os.IsNotExist(err) {
			usingConfigDefaults = true // no .conf file, so using defaults
			c = Config{}
			setDefaults(&c)
		} else {
			panic(fmt.Sprintf("Failed to load config file %s: %v",
				*configFile, err))
		}
	}

	path := c.LogPath
	if path == "" {
		path = defaultLogFile
	}

	if logFile != nil {
		path = *logFile
	}
	log.SetOutput(&lumberjack.Logger{
		Filename: path,
		MaxSize:  c.MaxLogSizeMb,  // megabytes
		MaxAge:   c.MaxLogAgeDays, // days
	})

	if *logLevel == "" {
		*logLevel = c.LogLevel
	}

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	log.SetFormatter(new(log_formatter.VmwareFormatter))
	log.SetLevel(level)

	if usingConfigDefaults {
		log.Info("No config file found. Using defaults.")
	}
	return usingConfigDefaults
}

// InitConfig set up driver specific options
func InitConfig(defaultConfigPath string, defaultLogPath string, defaultDriver string,
	defaultWindowsDriver string) (Config, error) {
	// Get options from ENV (where available), and from command line.
	// ENV takes precedence, so we can modify it in Docker plugin install
	logEnv := os.Getenv("VDVS_LOG_LEVEL")
	logLevel := &logEnv
	if *logLevel == "" {
		logLevel = flag.String("log_level", "", "Logging Level")
	}
	configFile := flag.String("config", defaultConfigPath, "Configuration file path")
	driverName := flag.String("driver", "", "Volume driver")

	// Photon driver options
	targetURL := flag.String("target", "", "Photon controller URL")
	projectID := flag.String("project", "", "Project ID of the docker host")
	vmID := flag.String("host", "", "ID of docker host")

	// vSphere driver options
	port := flag.Int("port", defaultPort, "Default port to connect to ESX service")
	useMockEsx := flag.Bool("mock_esx", false, "Mock the ESX service")

	flag.Parse()

	// Load the configuration if one was provided.
	c, err := load(*configFile)
	if err != nil {
		log.Warningf("Failed to load config file %s: %v", *configFile, err)
	}

	LogInit(logLevel, nil, defaultLogPath, configFile)

	// If no driver provided on the command line, use the one in the
	// config file or the default.
	if *driverName != "" {
		c.Driver = *driverName
	} else if c.Driver == "" {
		c.Driver = defaultDriver
	}

	// The windows plugin only supports the vsphere driver.
	if runtime.GOOS == "windows" && c.Driver != defaultWindowsDriver {
		msg := fmt.Sprintf("Plugin only supports the %s driver on Windows, ignoring parameter driver = %s.",
			defaultWindowsDriver, c.Driver)
		log.Warning(msg)
		fmt.Println(msg)
		c.Driver = defaultWindowsDriver
	}

	log.WithFields(log.Fields{
		"driver":    c.Driver,
		"log_level": *logLevel,
		"config":    *configFile,
	}).Info("Starting plugin ")

	if c.Driver == PhotonDriver && err == nil {
		if *targetURL != "" {
			c.Target = *targetURL
		}
		if *projectID != "" {
			c.Project = *projectID
		}
		if *vmID != "" {
			c.Host = *vmID
		}

		log.WithFields(log.Fields{
			"target":  c.Target,
			"project": c.Project,
			"host":    c.Host}).Info("Plugin options - ")

		if c.Target == "" || c.Project == "" || c.Host == "" {
			log.Warning("Invalid options specified for target/project/host")
			fmt.Printf("Invalid options specified for target - %s project - %s host - %s. Exiting.\n",
				c.Target, c.Project, c.Host)
			os.Exit(1)
		}
	} else if c.Driver == VSphereDriver || c.Driver == VMDKDriver {
		if c.Driver == VMDKDriver {
			log.Warning("Using deprecated \"vmdk\" driver, use \"vsphere\" driver instead - continuing...")
			c.Driver = VSphereDriver
		}

		c.Port = *port
		c.UseMockEsx = *useMockEsx

		log.WithFields(log.Fields{
			"port":       c.UseMockEsx,
			"useMockEsx": c.UseMockEsx}).Info("Plugin options - ")

	}

	return c, nil

}
