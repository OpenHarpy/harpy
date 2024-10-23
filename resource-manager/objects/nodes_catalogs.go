package objects

type NodeCatalog struct {
	// The Node Catalog is a struct that holds all the information about the nodes
	// It is used to store all the information about the nodes in the system
	NodeType         string `gorm:"primary_key"`
	NumCores         int
	AmountOfMemory   int
	AmountOfStorage  int
	NodeMaxCapacity  int
	NodeWarmpoolSize int
	NodeIdleTimeout  int
}

func (nc *NodeCatalog) Sync() { SyncGenericStruct(nc) }

// Getter functions for NodeCatalog
func NodeCatalogs() []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog
	result := db.Find(&ncs)
	if result.Error != nil {
		return nil
	}
	return ncs
}
func CatalogsWithWhereClause(whereClause string, args ...interface{}) []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog
	result := db.Where(whereClause, args...).Find(&ncs)
	if result.Error != nil {
		return nil
	}
	return ncs
}

func GetNodeCatalog(nodeType string) (*NodeCatalog, bool) {
	result := CatalogsWithWhereClause("node_type = ?", nodeType)
	if len(result) == 0 {
		return nil, false
	}
	return result[0], true
}

func CatalogsWithWarmPoolNotMet() []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog

	// Subquery to count live nodes per node type
	subQuery := db.Model(&LiveNode{}).
		Select("node_type, COUNT(*) as live_node_count").
		Group("node_type")

	// Join NodeCatalog with the subquery and filter based on NodeWarmpoolSize
	result := db.Table("node_catalogs").
		Select("node_catalogs.*").
		Joins("LEFT JOIN (?) AS live_nodes ON node_catalogs.node_type = live_nodes.node_type", subQuery).
		Where("node_catalogs.node_warmpool_size > IFNULL(live_nodes.live_node_count, 0)").
		Find(&ncs)

	if result.Error != nil {
		return nil
	}
	return ncs
}
