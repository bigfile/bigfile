//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"errors"
	"io"
	"os"

	"github.com/jinzhu/gorm"
)

var (
	// ErrInvalidObject represent a invalid object
	ErrInvalidObject = errors.New("invalid object")
	// ErrObjectNoChunks represent that a object has no any chunks
	ErrObjectNoChunks = errors.New("object has no any chunks")
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
}

// NewObjectReader is used to create a reader that read data from underlying chunk
func NewObjectReader(object *Object, rootPath *string, db *gorm.DB) (io.Reader, error) {

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
	if len(p) == 0 {
		return
	}
	if readCount, err = or.currentChunkReader.Read(p); err != nil {
		if err == io.EOF {
			_ = or.currentChunkReader.Close()
			if or.currentChunkNumber == or.totalChunkNumber {
				return readCount, io.EOF
			}
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
	return readCount, nil
}
