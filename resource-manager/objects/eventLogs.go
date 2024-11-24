package objects

type EventLogEntry struct {
	// This struct is used to store the configuration of the resource manager
	// It is used to store the configuration of the resource manager
	EventLogID   string `gorm:"primary_key"`
	EventLogType string `gorm:"type:text"`
	EventLogJSON string `gorm:"type:text"`
	EventGroupID string `gorm:"type:text"`
	EventLogTime int64
}

func (c *EventLogEntry) Sync()   { SyncGenericStruct(c) }
func (c *EventLogEntry) Delete() { DeleteGenericStruct(c) }

func GetEventLogs() ([]*EventLogEntry, bool) {
	db := GetDBInstance().db
	var els []*EventLogEntry
	result := db.Find(&els)
	if result.Error != nil {
		return nil, false
	}
	return els, false
}
func GetEventLogsByType(eventLogType string, page int, pageSize int) ([]*EventLogEntry, bool) {
	db := GetDBInstance().db
	var els []*EventLogEntry
	result := db.Where("event_log_type = ?", eventLogType).Find(&els)
	if result.Error != nil {
		return nil, false
	}
	return els, false
}

func CleanupEventLog(limit int) {
	// Cleanup the event log
	db := GetDBInstance().db
	db.Exec("DELETE FROM event_log_entries WHERE event_log_id NOT IN (SELECT event_log_id FROM event_log_entries ORDER BY event_log_time DESC LIMIT ?)", limit)
}
