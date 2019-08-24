//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"errors"
	"io"
	"os"
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
	object *Object

	rootPath           *string
	currentChunkIndex  int
	currentChunkReader *os.File
}

// NewObjectReader is used to create a reader that read data from underlying chunk
func NewObjectReader(object *Object, rootPath *string) (io.Reader, error) {

	if object == nil {
		return nil, ErrInvalidObject
	}

	if len(object.Chunks) == 0 {
		return nil, ErrObjectNoChunks
	}

	var (
		err         error
		chunkReader *os.File
	)

	if chunkReader, err = object.Chunks[0].Reader(rootPath); err != nil {
		return nil, err
	}

	return &objectReader{
		object:             object,
		currentChunkIndex:  0,
		currentChunkReader: chunkReader,
		rootPath:           rootPath,
	}, nil
}

func (or *objectReader) Read(p []byte) (int, error) {

	if len(p) == 0 {
		return 0, nil
	}

	var (
		err       error
		readCount int
	)

	if readCount, err = or.currentChunkReader.Read(p); err != nil {
		if err == io.EOF {
			if or.currentChunkIndex == len(or.object.Chunks)-1 {
				return readCount, io.EOF
			}
			_ = or.currentChunkReader.Close()
			or.currentChunkIndex++
			if or.currentChunkReader, err = or.object.Chunks[or.currentChunkIndex].Reader(or.rootPath); err != nil {
				return readCount, err
			}
			return readCount, nil
		}
		return readCount, err
	}

	return readCount, nil
}
