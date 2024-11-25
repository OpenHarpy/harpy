package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
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
		var usingConfigFileLocation string = "../default/global_config.json"
		// First we check if the enviroment variable is set
		configFileLocation, ok := os.LookupEnv("GLOBAL_CONFIG_FILE")
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

func (c *Configs) GetConfigIntWithDefault(ks string, defaultVal int) int {
	value, ok := c.Configs[ks]
	if !ok {
		return defaultVal
	}
	// Convert the string to an int
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return intValue
}

func (c *Configs) GetConfigWithDefault(ks string, defaultVal string) string {
	value, ok := c.Configs[ks]
	if !ok {
		return defaultVal
	}
	return value
}

func (c *Configs) GetInstanceID() string {
	return c.InstanceID
}

func (c *Configs) ValitateRequiredConfigs(keys []string) error {
	missing_configs := []string{}
	for _, key := range keys {
		_, ok := c.Configs[key]
		if !ok {
			missing_configs = append(missing_configs, key)
		}
	}
	if len(missing_configs) > 0 {
		return errors.New("Missing required configs: " + strings.Join(missing_configs, ", "))
	}
	return nil
}

func (c *Configs) GetAllConfigsWithPrefix(prefix string) map[string]string {
	if prefix == "" {
		return c.Configs
	}
	configs := make(map[string]string)
	for key, value := range c.Configs {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			configs[key] = value
		}
	}
	return configs
}
