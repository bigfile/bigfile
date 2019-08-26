//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	janitor "github.com/json-iterator/go"
)

func assertTokenRespStructure(data interface{}) bool {
	keys := []string{"availableTimes", "token", "ip", "readOnly", "expiredAt", "path", "secret"}
	mData := data.(map[string]interface{})
	for _, k := range keys {
		if _, ok := mData[k]; !ok {
			return false
		}
	}
	return true
}

func parseResponse(body string) (*Response, error) {
	var response = &Response{}
	json := janitor.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal([]byte(body), response)
	return response, err
}

// signRequestParams calculate the signature of params
func getParamsSignBody(p map[string]interface{}, secret string) string {
	sign, buf := getParamsSignatureWithBuf(p, secret)
	buf.WriteString(fmt.Sprintf("&sign=%s", sign))
	return buf.String()
}

// getParamsSignatureWithBuf
func getParamsSignatureWithBuf(p map[string]interface{}, secret string) (string, *strings.Builder) {
	var (
		buf   = new(strings.Builder)
		index = 0
		keys  = make([]string, len(p))
	)
	for k := range p {
		keys[index] = k
		index++
	}
	sort.Strings(keys)
	for i, k := range keys {
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteString(fmt.Sprintf("%v", p[k]))
		if i != len(keys)-1 {
			buf.WriteRune('&')
		}
	}
	return SignStrWithSecret(buf.String(), secret), buf
}

func getParamsSignature(p map[string]interface{}, secret string) string {
	sign, _ := getParamsSignatureWithBuf(p, secret)
	return sign
}

// fileChunk is used to split the whole file to chunks
func fileChunk(f string, size int) error {
	var (
		err  error
		file *os.File
	)

	if file, err = os.Open(f); err != nil {
		return err
	}

	for index := 0; ; index++ {
		var (
			chunk     = make([]byte, size)
			readCount int
			finished  bool
		)
		if readCount, err = file.Read(chunk); err != nil {
			if err != io.EOF {
				return err
			}
			finished = true
		}
		if readCount > 0 {
			if err = ioutil.WriteFile(fmt.Sprintf("%d-%s", index, f), chunk[:readCount], 0644); err != nil {
				return err
			}
		}
		if finished {
			break
		}
	}
	return nil
}
