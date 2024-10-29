package task

type BlockInternalReference struct {
	BlockID    string
	BlockType  string
	BlockGroup string
}

type BlockTracker struct {
	Blocks      []*BlockInternalReference
	BlockGroups map[string][]*BlockInternalReference
}

func NewBlockTracker() *BlockTracker {
	return &BlockTracker{
		Blocks:      make([]*BlockInternalReference, 0),
		BlockGroups: make(map[string][]*BlockInternalReference),
	}
}

func (bt *BlockTracker) AddBlock(block *BlockInternalReference) {
	bt.Blocks = append(bt.Blocks, block)
	if _, exists := bt.BlockGroups[block.BlockGroup]; !exists {
		bt.BlockGroups[block.BlockGroup] = make([]*BlockInternalReference, 0)
	}
	bt.BlockGroups[block.BlockGroup] = append(bt.BlockGroups[block.BlockGroup], block)
}

func (bt *BlockTracker) GetBlockGroup(blockGroupID string, filterOutBlockTypes []string) []*BlockInternalReference {
	if blockGroup, exists := bt.BlockGroups[blockGroupID]; exists {
		if filterOutBlockTypes != nil {
			blockTypeSet := make(map[string]struct{}, len(filterOutBlockTypes))
			for _, blockType := range filterOutBlockTypes {
				blockTypeSet[blockType] = struct{}{}
			}

			filteredBlockGroup := make([]*BlockInternalReference, 0)
			for _, block := range blockGroup {
				if _, found := blockTypeSet[block.BlockType]; found {
					continue
				} else {
					filteredBlockGroup = append(filteredBlockGroup, block)
				}
			}
			return filteredBlockGroup
		}
		return blockGroup
	}
	return nil
}
