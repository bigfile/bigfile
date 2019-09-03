//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package log

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	var (
		logConfig   = *config.DefaultConfig
		logger      *logging.Logger
		err         error
		tempLogFile *os.File
		wg          sync.WaitGroup
	)

	logConfig.Console.Enable = false
	tempLogFile, err = ioutil.TempFile(os.TempDir(), "*.log")
	assert.Equal(t, err, nil)

	defer os.Remove(tempLogFile.Name())
	logConfig.File.Path = tempLogFile.Name()

	logger, err = NewLogger(&logConfig.Log)
	assert.Equal(t, err, nil)

	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				logger.Error("hello world")
			}
		}()
	}

	wg.Wait()
}

func TestMustNewLogger(t *testing.T) {
	defer func() { assert.Nil(t, recover()) }()
	var (
		logConfig   = *config.DefaultConfig
		err         error
		tempLogFile *os.File
	)

	logConfig.Console.Enable = false
	tempLogFile, err = ioutil.TempFile(os.TempDir(), "*.log")
	assert.Equal(t, err, nil)

	defer os.Remove(tempLogFile.Name())
	logConfig.File.Path = tempLogFile.Name()

	MustNewLogger(&logConfig.Log)
}
