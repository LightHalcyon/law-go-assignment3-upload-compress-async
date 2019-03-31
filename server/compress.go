package server

import (
	"fmt"
	"io/ioutil"
	"io"
	"net/http"
	"os"
	"math"
	"strconv"
	"archive/zip"
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