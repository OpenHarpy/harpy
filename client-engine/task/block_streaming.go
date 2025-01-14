// Package task implements any task related operations
//
// This file contains the implementation of the BlockStreamingWriter and BlockStreamingReader structures
// These structures are used to stream blocks between the client and the node
// The BlockStreamingWriter is used to send blocks to the node
// The BlockStreamingReader is used to receive blocks from the node
//

package task

import (
	pb "client-engine/grpc_node_protocol"
	"client-engine/logger"
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
)

const (
	BLOCK_READER_BUFFER_SIZE = 1024 * 1024 // 1MB
)

/* Node */
type BlockReaderStatus int
type BlockWriterStatus int

const (
	// Block reader status
	BlockReaderStatusOpen   BlockReaderStatus = 0
	BlockReaderStatusClosed BlockReaderStatus = 1
	BlockReaderStatusError  BlockReaderStatus = 2
	// Block writter status
	BlockWriterStatusOpen   BlockWriterStatus = 0
	BlockWriterStatusClosed BlockWriterStatus = 1
	BlockWriterStatusError  BlockWriterStatus = 2
)

type BlockStreamingWriter struct {
	Status        BlockWriterStatus
	StreamHandler grpc.ClientStreamingClient[pb.BlockChunk, pb.BlockHandler]
	BlockID       string
}

type BlockStreamingReader struct {
	BlockID    string
	Status     BlockReaderStatus
	BufferFIFO [][]byte
	Buffer     []byte
}

func NewBlockStreamingWriter(n *Node) *BlockStreamingWriter {
	stream, err := n.client.StreamInBlock(context.Background())
	if err != nil {
		logger.Error("Failed to initiate stream", "NODE", err)
		return nil
	}
	return &BlockStreamingWriter{
		Status:        BlockWriterStatusOpen,
		StreamHandler: stream,
	}
}

func (b *BlockStreamingWriter) Write(buffer []byte) error {
	if b.Status != BlockWriterStatusOpen {
		return errors.New("BlockReader is not open")
	}
	err := b.StreamHandler.Send(&pb.BlockChunk{BlockChunk: buffer})
	if err != nil {
		logger.Error("Failed to send block chunk", "BLOCK-STREAM-WRITER", err)
		b.Status = BlockWriterStatusError
		return err
	}
	return nil
}

func (b *BlockStreamingWriter) Close() error {
	if b.Status != BlockWriterStatusOpen {
		return errors.New("BlockReader is not open")
	}
	response, err := b.StreamHandler.CloseAndRecv()
	if err != nil {
		logger.Error("Failed to close stream", "NODE", err)
		b.Status = BlockWriterStatusError
		return err
	}
	b.Status = BlockWriterStatusClosed
	b.BlockID = response.BlockID
	return nil
}

func BlockStreamInRoutine(sr *BlockStreamingReader, n *Node) {
	stream, err := n.client.StreamOutBlock(context.Background(), &pb.BlockHandler{BlockID: sr.BlockID})
	if err != nil {
		logger.Error("Failed to initiate stream", "BLOCK-STREAM-READER", err)
		sr.Status = BlockReaderStatusError
		return
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Error("Failed to receive block chunk", "BLOCK-STREAM-READER", err)
			sr.Status = BlockReaderStatusError
			break
		}
		sr.Buffer = append(sr.Buffer, chunk.BlockChunk...)
		if len(sr.Buffer) > BLOCK_READER_BUFFER_SIZE {
			// We need to flush the buffer
			sr.BufferFIFO = append(sr.BufferFIFO, sr.Buffer)
			sr.Buffer = []byte{}
		}
	}
	if len(sr.Buffer) > 0 {
		sr.BufferFIFO = append(sr.BufferFIFO, sr.Buffer)
	}
	sr.Status = BlockReaderStatusClosed
}

func NewBlockStreamingReader(blockID string, n *Node) *BlockStreamingReader {
	sr := &BlockStreamingReader{
		BlockID: blockID,
		Status:  BlockReaderStatusOpen,
		Buffer:  []byte{},
	}
	go BlockStreamInRoutine(sr, n)
	return sr
}

func (sr *BlockStreamingReader) Read() ([]byte, bool) {
	if sr.Status == BlockReaderStatusError {
		return nil, true
	} else if sr.Status == BlockReaderStatusClosed && len(sr.BufferFIFO) == 0 {
		return nil, true
	} else if len(sr.BufferFIFO) == 0 {
		return nil, false
	}
	buffer := sr.BufferFIFO[0]
	sr.BufferFIFO = sr.BufferFIFO[1:]
	return buffer, false
}
