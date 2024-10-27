package main

import (
	"errors"
	"os"
	"remote-runner/config"
	"strings"
)

const BlockBufferSizeInBytes = 1024 * 1024 * 4 // 4MB
const BlockChunkSizeInBytes = 1024 * 1024      // 1MB

type BlockWriterState uint32
type BlockReaderState uint32

const (
	// Block states
	BlockWriterStateNew       BlockWriterState = 0
	BlockWriterStateStreaming BlockWriterState = 1
	BlockWriterStateSealed    BlockWriterState = 2
	// Block reader states
	BlockReaderStateNew       BlockReaderState = 0
	BlockReaderStateStreaming BlockReaderState = 1
	BlockReaderStateStreamEnd BlockReaderState = 2
)

// ** Block ** //
func ClearAllBlocksInDir() error {
	blockLocationRoot := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.blockMountLocation", "./_blocks")
	files, err := os.ReadDir(blockLocationRoot)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := os.Remove(blockLocationRoot + "/" + file.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

type Block struct {
	BlockID       string
	BlockLocation string
}

func (b *Block) CheckBlockExists() bool {
	_, err := os.Stat(b.BlockLocation)
	return err == nil
}

func (b *Block) Cleanup() error {
	if b.BlockLocation == "" {
		return errors.New("block location is empty")
	} else if b.BlockLocation == "/" {
		return errors.New("block location is root")
	} else if b.BlockLocation == "." {
		return errors.New("block location is current directory")
	} else if b.BlockLocation == ".." {
		return errors.New("block location is parent directory")
	} else if b.BlockLocation == "./" {
		return errors.New("block location is current directory")
	} else if b.BlockLocation == "../" {
		return errors.New("block location is parent directory")
	}
	if b.CheckBlockExists() {
		err := os.Remove(b.BlockLocation)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewBlock(blockID string) *Block {
	blockLocationRoot := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.blockMountLocation", "./_blocks")
	blockFilePath := strings.ReplaceAll(blockLocationRoot+"/"+blockID+".block", "//", "/")
	return &Block{
		BlockID:       blockID,
		BlockLocation: blockFilePath,
	}
}

func WriteBytesToBlock(data []byte, block *Block) error {
	// Write the data to the block location
	file, err := os.Create(block.BlockLocation)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// ** Block Reader ** //
type BlockReader struct {
	Block      *Block
	IOHandler  *os.File
	Index      int
	State      BlockReaderState
	BlockEndAt int
}

func (br *BlockReader) ReadChunk() ([]byte, error) {
	if br.State != BlockReaderStateStreaming {
		return nil, errors.New("block reader is not in streaming state")
	}
	// Read the chunk from the block
	blockStart := br.Index * BlockChunkSizeInBytes
	blockEnd := (br.Index + 1) * BlockChunkSizeInBytes
	if blockEnd > br.BlockEndAt {
		blockEnd = br.BlockEndAt
	}

	chunk := make([]byte, blockEnd-blockStart)
	_, err := br.IOHandler.ReadAt(chunk, int64(blockStart))
	if err != nil {
		return nil, err
	}
	br.Index++
	if blockEnd == br.BlockEndAt {
		br.State = BlockReaderStateStreamEnd
		// Close the file
		err = br.IOHandler.Close()
		if err != nil {
			return nil, err
		}
	}
	return chunk, nil
}

func (br *BlockReader) StartReading() error {
	// Open the file
	file, err := os.Open(br.Block.BlockLocation)
	if err != nil {
		return err
	}
	br.IOHandler = file
	br.State = BlockReaderStateStreaming
	blockEndAt, err := br.IOHandler.Seek(0, 2)
	if err != nil {
		return err
	}
	br.BlockEndAt = int(blockEndAt)
	return nil
}

func NewBlockReader(block *Block) (*BlockReader, error) {
	blockReader := &BlockReader{
		Block:      block,
		IOHandler:  nil,
		Index:      0,
		State:      BlockReaderStateNew,
		BlockEndAt: 0,
	}
	err := blockReader.StartReading()
	return blockReader, err
}

// ** Block Reader ** //
type BlockWriter struct {
	Block       *Block
	BlockBuffer []byte
	BlockState  BlockWriterState
	IOHandler   *os.File
}

func (bw *BlockWriter) StartWriting() error {
	// Create the file
	file, err := os.Create(bw.Block.BlockLocation)
	if err != nil {
		return err
	}
	bw.IOHandler = file
	bw.BlockState = BlockWriterStateStreaming
	return nil
}

func (bw *BlockWriter) FlushBlock() error {
	if bw.BlockState != BlockWriterStateStreaming {
		return errors.New("block is not in streaming state")
	}
	if len(bw.BlockBuffer) == 0 {
		return nil
	}
	// Write the block buffer to the block location
	_, err := bw.IOHandler.Write(bw.BlockBuffer)
	if err != nil {
		return err
	}
	// Clear the block buffer
	bw.BlockBuffer = []byte{}
	return nil
}

func (bw *BlockWriter) SealBlock() error {
	// Flush the block buffer
	err := bw.FlushBlock()
	if err != nil {
		return err
	}
	// Close the file
	err = bw.IOHandler.Close()
	if err != nil {
		return err
	}
	bw.BlockState = BlockWriterStateSealed
	return nil
}

func (bw *BlockWriter) AppendChunk(chunk []byte) {
	// Append the chunk to the block buffer
	bw.BlockBuffer = append(bw.BlockBuffer, chunk...)
	// If the block buffer is full, write the block to the block location
	if len(bw.BlockBuffer) >= BlockBufferSizeInBytes {
		bw.FlushBlock()
	}
}

func NewBlockWriter(block *Block) (*BlockWriter, error) {
	blockWriter := &BlockWriter{
		Block:       block,
		BlockBuffer: []byte{},
		BlockState:  BlockWriterStateNew,
		IOHandler:   nil,
	}
	err := blockWriter.StartWriting()
	return blockWriter, err
}
