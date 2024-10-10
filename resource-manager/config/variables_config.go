package config

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
)

type Configs struct {
	InstanceID string
	Configs    map[string]string
}

var instance *Configs
var once sync.Once

func GetConfigs() *Configs {
	once.Do(func() {
		var instanceID = uuid.New().String()
		var options map[string]string
		var usingConfigFileLocation string = "config.json"
		// First we check if the enviroment variable is set
		configFileLocation, ok := os.LookupEnv("RESOURCE_MANAGER_CONFIG_FILE")
		if !ok {
			println("(PRELOGGER) [WARNING] - No config file location set, using default location")
		} else {
			usingConfigFileLocation = configFileLocation
		}

		configFileContents, err := os.ReadFile(usingConfigFileLocation)
		if err != nil {
			println("(PRELOGGER) [ERROR] - Error reading config file", err)
			os.Exit(1)
		}

		err = json.Unmarshal(configFileContents, &options)
		if err != nil {
			println("(PRELOGGER) [ERROR] - Error unmarshalling config file", err)
			os.Exit(1)
		}

		// Print all configs
		for key, value := range options {
			println("(PRELOGGER) [INFO] - Config Key: ", key, " Config Value: ", value)
		}

		instance = &Configs{
			InstanceID: instanceID,
			Configs:    options,
		}
	})
	return instance
}

// Create a function on the Configs to get a specific config
func (c *Configs) GetConfig(ks string) (string, bool) {
	value, ok := c.Configs[ks]
	return value, ok
}

func (c *Configs) GetConfigsWithDefault(ks string, defaultVal string) string {
	value, ok := c.Configs[ks]
	if !ok {
		return defaultVal
	}
	return value
}

func (c *Configs) GetInstanceID() string {
	return c.InstanceID
}
