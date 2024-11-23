package objects

type Config struct {
	// This struct is used to store the configuration of the resource manager
	// It is used to store the configuration of the resource manager
	ConfigKey   string `gorm:"primary_key"`
	ConfigValue string `gorm:"type:text"`
}

func (c *Config) Sync()   { SyncGenericStruct(c) }
func (c *Config) Delete() { DeleteGenericStruct(c) }

// Getter functions for Config
func GetConfig(configKey string) (*Config, bool) {
	db := GetDBInstance().db
	var c Config
	result := db.First(&c, "config_key = ?", configKey)
	if result.Error != nil {
		return nil, false
	}
	return &c, true
}

func GetConfigs() []*Config {
	db := GetDBInstance().db
	var cs []*Config
	result := db.Find(&cs)
	if result.Error != nil {
		return nil
	}
	return cs
}
