//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/jinzhu/gorm"
)

var (
	// ErrInvalidObject represent a invalid object
	ErrInvalidObject = errors.New("invalid object")
	// ErrObjectNoChunks represent that a object has no any chunks
	ErrObjectNoChunks = errors.New("object has no any chunks")
	// ErrInvalidSeekWhence represent invalid seek whence
	ErrInvalidSeekWhence = errors.New("invalid seek whence")
	// ErrNegativePosition represent negative position
	ErrNegativePosition = errors.New("negative read position")
)

// ObjectReader is used to read data from underlying chunk.
// until read all data, it will return io.EOF
type objectReader struct {
	db                 *gorm.DB
	object             *Object
	rootPath           *string
	currentChunkReader *os.File
	totalChunkNumber   int
	currentChunkNumber int
	alreadyReadCount   int
}

// NewObjectReader is used to create a reader that read data from underlying chunk
func NewObjectReader(object *Object, rootPath *string, db *gorm.DB) (io.ReadSeeker, error) {

	if object == nil {
		return nil, ErrInvalidObject
	}

	var (
		err              error
		firstChunk       *Chunk
		chunkReader      *os.File
		totalChunkNumber int
	)

	if totalChunkNumber, err = object.LastChunkNumber(db); err != nil {
		return nil, err
	}

	if totalChunkNumber == 0 {
		return nil, ErrObjectNoChunks
	}

	if firstChunk, err = object.ChunkWithNumber(1, db); err != nil {
		return nil, err
	}

	if chunkReader, err = firstChunk.Reader(rootPath); err != nil {
		return nil, err
	}

	return &objectReader{
		db:                 db,
		object:             object,
		currentChunkReader: chunkReader,
		rootPath:           rootPath,
		currentChunkNumber: 1,
		totalChunkNumber:   totalChunkNumber,
	}, nil
}

func (or *objectReader) Read(p []byte) (readCount int, err error) {
	if or.alreadyReadCount >= or.object.Size {
		_ = or.currentChunkReader.Close()
		return 0, io.EOF
	}
	defer func() { or.alreadyReadCount += readCount }()
	readCount, err = or.currentChunkReader.Read(p)
	if err != nil && err == io.EOF {
		_ = or.currentChunkReader.Close()
		or.currentChunkNumber++
		var nextChunk *Chunk
		if nextChunk, err = or.object.ChunkWithNumber(or.currentChunkNumber, or.db); err != nil {
			return
		}
		if or.currentChunkReader, err = nextChunk.Reader(or.rootPath); err != nil {
			return readCount, err
		}
		return readCount, nil
	}
	return readCount, err
}

func (or *objectReader) Seek(offset int64, whence int) (abs int64, err error) {
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(or.alreadyReadCount) + offset
	case io.SeekEnd:
		abs = int64(or.object.Size) + offset
	default:
		return 0, ErrInvalidSeekWhence
	}
	if abs < 0 {
		return 0, ErrNegativePosition
	}
	if abs >= int64(or.object.Size) {
		or.alreadyReadCount = int(abs)
		or.currentChunkNumber = or.totalChunkNumber
		return abs, nil
	}
	var (
		currentChunk       *Chunk
		currentChunkReader *os.File
		currentChunkNumber = int(math.Ceil(float64(abs) / float64(ChunkSize)))
	)

	if abs%ChunkSize == 0 {
		currentChunkNumber++
	}

	if currentChunkNumber == or.currentChunkNumber {
		currentChunkReader = or.currentChunkReader
	} else {
		if currentChunk, err = or.object.ChunkWithNumber(currentChunkNumber, or.db); err != nil {
			return 0, nil
		}
		if currentChunkReader, err = currentChunk.Reader(or.rootPath); err != nil {
			return 0, err
		}
	}
	if _, err = currentChunkReader.Seek(abs%ChunkSize, io.SeekStart); err != nil {
		return 0, err
	}
	if currentChunkNumber != or.currentChunkNumber {
		_ = or.currentChunkReader.Close()
	}
	or.currentChunkReader = currentChunkReader
	or.currentChunkNumber = currentChunkNumber
	or.alreadyReadCount = int(abs)
	return abs, nil
}
