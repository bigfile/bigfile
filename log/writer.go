//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package log plan to provide a log collect component that can
// output application log to console and file meanwhile. In addition,
// log that is exported to file can be rotated automatically by Mode
// and file size.
package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// AutoRotateWriter implement io.Writer interface, this writer will
// write content to different file by configuration.
type AutoRotateWriter struct {
	mu sync.Mutex

	// maxBytes represent single log file max size
	maxBytes uint64

	// dir represent the directory that include log file
	dir string

	// number represent log number by mode, the complete log file
	// name calculated by ${parent}/${basename}.${number}.${ext}
	number uint32

	// basename represent log file basename, example: bigfile
	basename string

	// ext log file ext
	ext string

	// handler represent current file descriptor, that's used
	// to write log content
	handler *os.File

	// handlerAlreadyWriteSize represent the size current handler have
	// already wrote to log file
	handlerAlreadyWriteSize uint64
}

// NewAutoRotateWriter is used return a writer, that will change to another writer
// when the size of current file will be up to maxBytes.
func NewAutoRotateWriter(file string, maxBytes uint64) (*AutoRotateWriter, error) {
	var (
		dir                     = filepath.Dir(file)
		err                     error
		stat                    os.FileInfo
		number                  uint32
		handlerAlreadyWriteSize uint64
		completeFileName        string
		basename                string
		ext                     string
		handler                 *os.File
	)
	stat, err = os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	} else if err == nil && !stat.IsDir() {
		return nil, errors.New("the directory of log file is illegal   ")
	} else if err != nil {
		return nil, err
	}

	dir = strings.TrimSuffix(dir, "/")
	ext = filepath.Ext(file)
	basename = strings.TrimSuffix(filepath.Base(file), ext)
	ext = strings.TrimPrefix(ext, ".")

	for {
		completeFileName = fmt.Sprintf("%s/%s.%d.%s", dir, basename, number, ext)
		if stat, err = os.Stat(completeFileName); err != nil {
			if os.IsNotExist(err) {
				if handler, err = os.OpenFile(
					completeFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
					return nil, err
				}
				break
			} else {
				return nil, err
			}
		} else {
			handlerAlreadyWriteSize = uint64(stat.Size())
			if handlerAlreadyWriteSize < maxBytes {
				handler, err = os.OpenFile(completeFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		number++
	}

	return &AutoRotateWriter{
		maxBytes:                maxBytes,
		dir:                     dir,
		number:                  number,
		basename:                basename,
		ext:                     ext,
		handler:                 handler,
		handlerAlreadyWriteSize: handlerAlreadyWriteSize,
	}, nil
}

// Write is used to implement io.Writer, because of mutex before write,
// this will lead to a bad performance. I will optimize this in the future.
// TODO: optimize performance
func (a *AutoRotateWriter) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	n, err = a.handler.Write(p)
	atomic.AddUint64(&a.handlerAlreadyWriteSize, uint64(n))
	if atomic.LoadUint64(&a.handlerAlreadyWriteSize) >= a.maxBytes {
		_ = a.handler.Close()
		atomic.AddUint32(&a.number, 1)
		nextFileName := fmt.Sprintf("%s/%s.%d.%s", a.dir, a.basename, a.number, a.ext)
		a.handler, err = os.OpenFile(nextFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return n, err
		}
		a.handlerAlreadyWriteSize = 0
	}
	return n, err
}

// Close is used to implement io.Close
func (a *AutoRotateWriter) Close() error {
	return a.handler.Close()
}
