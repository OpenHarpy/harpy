package chunker

const CHUNK_SIZE = 1024 * 1024 // 1MB

func ChunkBytes(data []byte, index *int) []byte {
	// This function will chunk the bytes into chunks of size chunkSize
	// It will return the chunked bytes and update the index
	// If the index is greater than the length of the data then it will return an empty byte array
	chunkSize := CHUNK_SIZE
	if *index >= len(data) {
		return []byte{}
	}
	if *index+chunkSize > len(data) {
		chunkSize = len(data) - *index
	}
	chunk := data[*index : *index+chunkSize]
	*index += chunkSize
	return chunk
}
