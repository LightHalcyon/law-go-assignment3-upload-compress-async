package main

import (
	"os"
	"math"
	"compress/gzip"
	"bytes"
)

// Split splits file to chunks
func Split(file os.File) [10][]byte {
	var chunks [10][]byte

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fileChunk := int64(math.Ceil(float64(fileSize) / float64(10)))

	for i := 0; i < 10; i++ {
		partSize := int(math.Min(float64(fileChunk), float64(fileSize-int64(int64(i)*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)

		chunks[i] = partBuffer
	}

	return chunks
}

// Compress compress bytes given
func Compress(file []byte) ([]byte, error) {
	var err error
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	
	_, err = gz.Write(file)
	if err != nil {
		err = gz.Flush()
	}
	if err != nil {
		err = gz.Close()
	}

	return b.Bytes(), err
}