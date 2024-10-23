package objects

type Provider struct {
	ProviderName        string `gorm:"primary_key"`
	ProviderDescription string `gorm:"type:text"`
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
func GetProviders() []*Provider {
	db := GetDBInstance().db
	var providers []*Provider
	result := db.Find(&providers)
	if result.Error != nil {
		return nil
	}
	return providers
}
