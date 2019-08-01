//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package sha512 provides functions to export the middle state of
// internal sha512.digest and set the state. It also provides functions
// to serialize and deserialize between text and sha512.digest.
package sha512

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"hash"
	"reflect"
	"unsafe"
)

// State is a representation of *sha512.digest
type State struct {

	// H is corresponding to sha512.digest.h
	H [8]uint64

	// X is corresponding to sha512.digest.x
	X [128]byte

	// Nx is corresponding to sha512.digest.nx
	Nx int

	// Len is corresponding to sha512.digest.len
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
// reflect sha512.digest type.
var ErrDigestType = errors.New("digest must be type of *sha512.digest")

// GetHashState will return sha512.digest internal state. This is an unsafe method,
// so you should use it with caution. If reflect successfully, it will return a *State.
func GetHashState(digest hash.Hash) (*State, error) {

	if reflect.TypeOf(digest).String() != "*sha512.digest" {
		return nil, ErrDigestType
	}

	digestElem := reflect.ValueOf(digest).Elem()

	var (
		h    [8]uint64
		x    [128]byte
		nx   int
		xLen uint64
	)
	// sha512.digest.h
	rfh := digestElem.FieldByName("h")
	rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	h = rfh.Interface().([8]uint64)

	// sha512.digest.x
	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	x = rfx.Interface().([128]byte)

	// sha512.digest.nx
	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	nx = rfnx.Interface().(int)

	// sha512.digest.len
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

// SetHashState will be used to set sha512.digest state. This method will help us
// implement continuous hash.
func SetHashState(digest hash.Hash, state *State) error {

	if reflect.TypeOf(digest).String() != "*sha512.digest" {
		return ErrDigestType
	}
	digestElem := reflect.ValueOf(digest).Elem()

	rfh := digestElem.FieldByName("h")
	rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	rfhp := (*[8]uint64)(unsafe.Pointer(rfh.UnsafeAddr()))
	*rfhp = state.H

	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	rfxp := (*[128]byte)(unsafe.Pointer(rfx.UnsafeAddr()))
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

// NewHashWithStateText is a helper method, directly generate *sha512.digest by stateCipherText.
// Maybe, it will raise an error, return value err will represent it.
// If successfully, digest returned will be available.
func NewHashWithStateText(stateCipherText string) (digest hash.Hash, err error) {
	state, err := DecodeStringToState(stateCipherText)
	if err != nil {
		return nil, err
	}
	digest = sha512.New()
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
