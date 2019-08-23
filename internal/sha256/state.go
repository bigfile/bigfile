//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package sha256 provides functions to export the middle state of
// internal sha256.digest and set the state. It also provides functions
// to serialize and deserialize between text and sha256.digest.
package sha256

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"hash"
	"reflect"
	"unsafe"
)

// State is a representation of *sha256.digest
type State struct {

	// H is corresponding to sha256.digest.h
	H [8]uint32

	// X is corresponding to sha256.digest.x
	X [64]byte

	// Nx is corresponding to sha256.digest.nx
	Nx int

	// Len is corresponding to sha256.digest.len
	Len uint64
}

// EncodeToString encodes state to string by base64 encode.
// If there are anything wrong, it will return an error to
// represent it.
func (s *State) EncodeToString() (string, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(s); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecodeStringToState decodes string that is encoded by base64,
// and then decodes it to a *state. When something goes wrong, it will
// return an error, otherwise err is nil.
func DecodeStringToState(cipherText string) (*State, error) {
	plainTextByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	state := &State{}
	buf := bytes.Buffer{}
	buf.Write(plainTextByte)
	decoder := gob.NewDecoder(&buf)
	if err = decoder.Decode(state); err != nil {
		return nil, err
	}
	return state, nil
}

// ErrDigestType hash.Hash has many implementation types, but, here, we only
// reflect sha256.digest type.
var ErrDigestType = errors.New("digest must be type of *sha256.digest")

// GetHashState will return sha256.digest internal state. This is an unsafe method,
// so you should use it with caution. If reflect successfully, it will return a *State.
func GetHashState(digest hash.Hash) (*State, error) {

	if reflect.TypeOf(digest).String() != "*sha256.digest" {
		return nil, ErrDigestType
	}

	digestElem := reflect.ValueOf(digest).Elem()

	var (
		h    [8]uint32
		x    [64]byte
		nx   int
		xLen uint64
	)
	// sha256.digest.h
	rfh := digestElem.FieldByName("h")
	for i := 0; i < rfh.Len(); i++ {
		h[i] = *((*uint32)(unsafe.Pointer(rfh.Index(i).UnsafeAddr())))
	}
	//rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	//h = rfh.Interface().([8]uint32)

	// sha256.digest.x
	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	x = rfx.Interface().([64]byte)

	// sha256.digest.nx
	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	nx = rfnx.Interface().(int)

	// sha256.digest.len
	rfxLen := digestElem.FieldByName("len")
	rfxLen = reflect.NewAt(rfxLen.Type(), unsafe.Pointer(rfxLen.UnsafeAddr())).Elem()
	xLen = rfxLen.Interface().(uint64)

	return &State{
		H:   h,
		X:   x,
		Nx:  nx,
		Len: xLen,
	}, nil
}

// SetHashState will be used to set sha256.digest state. This method will help us
// implement continuous hash.
func SetHashState(digest hash.Hash, state *State) error {

	if reflect.TypeOf(digest).String() != "*sha256.digest" {
		return ErrDigestType
	}
	digestElem := reflect.ValueOf(digest).Elem()

	rfh := digestElem.FieldByName("h")
	rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	rfhp := (*[8]uint32)(unsafe.Pointer(rfh.UnsafeAddr()))
	*rfhp = state.H

	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	rfxp := (*[64]byte)(unsafe.Pointer(rfx.UnsafeAddr()))
	*rfxp = state.X

	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	rfnxp := (*int)(unsafe.Pointer(rfnx.UnsafeAddr()))
	*rfnxp = state.Nx

	rfxLen := digestElem.FieldByName("len")
	rfxLen = reflect.NewAt(rfxLen.Type(), unsafe.Pointer(rfxLen.UnsafeAddr())).Elem()
	rfxLenP := (*uint64)(unsafe.Pointer(rfxLen.UnsafeAddr()))
	*rfxLenP = state.Len

	return nil
}

// NewHashWithStateText is a helper method, directly generate *sha256.digest by stateCipherText.
// Maybe, it will raise an error, return value err will represent it.
// If successfully, digest returned will be available.
func NewHashWithStateText(stateCipherText string) (digest hash.Hash, err error) {
	state, err := DecodeStringToState(stateCipherText)
	if err != nil {
		return nil, err
	}
	digest = sha256.New()
	if err = SetHashState(digest, state); err != nil {
		return nil, err
	}
	return digest, nil
}

// GetHashStateText is also a helper method, it will return text representation of hash or
// unexpected error.
func GetHashStateText(digest hash.Hash) (string, error) {
	state, err := GetHashState(digest)
	if err != nil {
		return "", err
	}
	return state.EncodeToString()
}
