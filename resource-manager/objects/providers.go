package objects

type Provider struct {
	ProviderName          string `gorm:"primary_key"`
	ProviderDescription   string `gorm:"type:text"`
	ProviderCorePerNode   int    `gorm:"type:int"`
	ProviderMemoryPerNode int    `gorm:"type:int"`
	ProviderWarmpoolSize  int    `gorm:"type:int"`
	ProviderMaxNodes      int    `gorm:"type:int"`
}

func (p *Provider) Sync()   { SyncGenericStruct(p) }
func (p *Provider) Delete() { DeleteGenericStruct(p) }

type ProviderProps struct {
	ProviderName string `gorm:"primary_key"`
	PropKey      string `gorm:"primary_key"`
	PropValue    string `gorm:"type:text"`
}

func (p *ProviderProps) Sync()   { SyncGenericStruct(p) }
func (p *ProviderProps) Delete() { DeleteGenericStruct(p) }

// Getter functions for Provider
func GetProvider() *Provider {
	db := GetDBInstance().db
	var providers []*Provider
	result := db.Find(&providers)
	if result.Error != nil {
		return nil
	}
	if len(providers) == 0 {
		return nil
	}
	return providers[0]
}

func ResetProvider() {
	GetProvider().Delete()
}
