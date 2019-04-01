package main

import (
	"math"
	"compress/gzip"
	"bytes"
)

// Split splits file to chunks
func Split(file []byte) [10][]byte {
	var chunks [10][]byte
	
	fileSize := int64(len(file))
	fileChunk := int64(math.Ceil(float64(fileSize) / float64(10)))

	for i := 0; i < 10; i++ {
		partSize := int(math.Min(float64(fileChunk), float64(fileSize-int64(i)*fileChunk)))
		chunks[i] = file[i*partSize:((i+1)*partSize)-1]
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

// Combine combines chunk into bytes
func Combine(cfiles [10][]byte) []byte {
	var cfile []byte
	for _, v := range cfiles {
		cfile = append(cfile, v...)
	}
	return cfile
}