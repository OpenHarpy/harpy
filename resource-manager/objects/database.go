package objects

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Type Marker Indicator
type TypeMarkerIndicator int

// DB Struct (Singleton)
type DBInstance struct {
	db *gorm.DB
}

var dbInstance *DBInstance
var dbOnce sync.Once

func databaseMain() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("resource-manager.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&NodeCatalog{})
	db.AutoMigrate(&LiveNode{})
	db.AutoMigrate(&ResourceAssignment{})
	db.AutoMigrate(&Config{})
	db.AutoMigrate(&Provider{})
	db.AutoMigrate(&ProviderProps{})
	return db
}

func GetDBInstance() *DBInstance {
	dbOnce.Do(func() {
		db := databaseMain()
		dbInstance = &DBInstance{
			db: db,
		}
	})
	return dbInstance
}

func SyncGenericStruct(data interface{}) {
	db := GetDBInstance().db
	db.Save(data)
}

func DeleteGenericStruct(data interface{}) {
	db := GetDBInstance().db
	db.Delete(data)
}

func EnumToString(enum interface{}) string {
	// If the interface is a LiveNodeStatusEnum
	if _, ok := enum.(LiveNodeStatusEnum); ok {
		reverseMapping := make(map[LiveNodeStatusEnum]string)
		for k, v := range NodeStatusMapping {
			reverseMapping[v] = k
		}
		return reverseMapping[enum.(LiveNodeStatusEnum)]
	} else if _, ok := enum.(ResourceAssignmentStatusEnum); ok {
		reverseMapping := make(map[ResourceAssignmentStatusEnum]string)
		for k, v := range ResourceAssignmentStatusMapping {
			reverseMapping[v] = k
		}
		return reverseMapping[enum.(ResourceAssignmentStatusEnum)]
	}
	return ""
}

func StringToEnum(enumType TypeMarkerIndicator, str string) interface{} {
	if str == "" {
		return nil
	}
	if enumType == TypeMarker_ResourceAssignmentStatusEnum {
		return ResourceAssignmentStatusMapping[str]
	} else if enumType == TypeMarker_LiveNodeStatusEnum {
		return NodeStatusMapping[str]
	}
	return nil
}
