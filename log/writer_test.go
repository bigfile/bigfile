//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package log

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAutoRotateWriter(t *testing.T) {
	var (
		writer     *AutoRotateWriter
		err        error
		randNum    int
		filePath   string
		writeBytes int
	)

	// directory exists
	filePath = filepath.Join(os.TempDir(), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	assert.Equal(t, writer.dir, strings.TrimSuffix(os.TempDir(), "/"))
	assert.Equal(t, writer.Close(), nil)
	assert.Equal(t, writer.ext, "log")
	assert.Equal(t, writer.number, uint32(0))
	defer os.Remove(filePath)

	// directory doesn't exists
	rand.Seed(time.Now().UnixNano())
	randNum = int(rand.Int31n(100000000))
	filePath = filepath.Join(os.TempDir(), strconv.Itoa(randNum), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	assert.Equal(t, writer.dir, filepath.Join(os.TempDir(), strconv.Itoa(randNum)))
	writeBytes, err = writer.Write([]byte("hello"))
	assert.Equal(t, writeBytes, 5)
	assert.Equal(t, err, nil)
	assert.Equal(t, writer.Close(), nil)
	defer os.RemoveAll(filepath.Dir(filePath))

	// file exists
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	assert.Equal(t, uint64(5), writer.handlerAlreadyWriteSize)
	writeBytes, err = writer.Write([]byte("hello"))
	assert.Equal(t, writeBytes, 5)
	assert.Equal(t, err, nil)
	assert.Equal(t, writer.Close(), nil)

	// test change file handler automatically
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	assert.Equal(t, uint64(0), writer.handlerAlreadyWriteSize)
	assert.Equal(t, uint32(1), writer.number)
	assert.Equal(t, writer.Close(), nil)
}

func TestAutoRotateWriter_Close(t *testing.T) {
	var (
		writer   *AutoRotateWriter
		err      error
		randNum  int
		filePath string
	)
	rand.Seed(time.Now().UnixNano())
	randNum = int(rand.Int31n(100000000))
	filePath = filepath.Join(os.TempDir(), strconv.Itoa(randNum), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	assert.Equal(t, writer.Close(), nil)
	defer os.RemoveAll(filepath.Dir(filePath))
}

func TestAutoRotateWriter_Write(t *testing.T) {
	var (
		writer      *AutoRotateWriter
		err         error
		randNum     int
		filePath    string
		handlerName string
		writeBytes  int
	)
	rand.Seed(time.Now().UnixNano())
	randNum = int(rand.Int31n(100000000))
	filePath = filepath.Join(os.TempDir(), strconv.Itoa(randNum), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	defer os.RemoveAll(filepath.Dir(filePath))
	handlerName = writer.handler.Name()
	writeBytes, err = writer.Write([]byte("hellohello"))
	assert.Equal(t, writeBytes, 10)
	assert.Equal(t, err, nil)

	_, err = writer.Write([]byte("hello"))
	assert.Equal(t, err, nil)
	assert.Equal(t, uint32(1), writer.number)
	assert.NotEqual(t, writer.handler.Name(), handlerName)

	assert.Equal(t, writer.Close(), nil)
}

func TestAutoRotateWriter_Write2(t *testing.T) {
	var (
		err        error
		randNum    int
		filePath   string
		wg         sync.WaitGroup
		writeBytes int
		writer     *AutoRotateWriter
	)
	// test write concurrently
	filePath = filepath.Join(os.TempDir(), strconv.Itoa(randNum), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	assert.Equal(t, err, nil)
	defer os.RemoveAll(filepath.Dir(filePath))

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				writeBytes, err = writer.Write([]byte("hello hello"))
				assert.Equal(t, err, nil)
				assert.Equal(t, writeBytes, 11)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkAutoRotateWriter_Write(b *testing.B) {
	var (
		err      error
		randNum  int
		filePath string
		writer   *AutoRotateWriter
	)
	filePath = filepath.Join(os.TempDir(), strconv.Itoa(randNum), "bigfile.log")
	writer, err = NewAutoRotateWriter(filePath, 10)
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(filePath))

	for i := 0; i < b.N; i++ {
		_, err = writer.Write([]byte("hello hello"))
		if err != nil {
			b.Fatal(err)
		}
	}
}
