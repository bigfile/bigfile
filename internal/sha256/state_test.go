//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package sha256

import (
	"bytes"
	"crypto/sha256"
	"reflect"
	"testing"
	"unsafe"
)

func TestGetHashState(t *testing.T) {
	h := sha256.New()
	hState, err := GetHashState(h)
	if err != nil {
		t.Fatal(err)
	}

	digestElem := reflect.ValueOf(h).Elem()

	// sha256.digest.h
	rfh := digestElem.FieldByName("h")
	rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	if v := rfh.Interface().([8]uint32); v != hState.H {
		t.Fatalf("hState.H should be %v", v)
	}

	// sha256.digest.x
	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	if v := rfx.Interface().([64]byte); v != hState.X {
		t.Fatalf("hState.X should be %v", v)
	}

	// sha256.digest.nx
	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	if nx := rfnx.Interface().(int); nx != hState.Nx {
		t.Fatalf("hState.Nx should be %v", nx)
	}

	// sha256.digest.len
	rfxLen := digestElem.FieldByName("len")
	rfxLen = reflect.NewAt(rfxLen.Type(), unsafe.Pointer(rfxLen.UnsafeAddr())).Elem()
	if len_ := rfxLen.Interface().(uint64); len_ != hState.Len {
		t.Fatalf("hState.Len should be %v", len_)
	}
}

func TestSetHashState(t *testing.T) {
	h := sha256.New()

	h1 := sha256.New()
	if _, err := h1.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(h.Sum(nil), h1.Sum(nil)) {
		t.Fatalf("h and h1 sum should not be eqal")
	}

	h1State, err := GetHashState(h1)
	if err != nil {
		t.Fatal(err)
	}

	if err := SetHashState(h, h1State); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(h.Sum(nil), h1.Sum(nil)) {
		t.Fatalf("h and h1 sum should be eqal")
	}
}

func TestState_EncodeToString(t *testing.T) {
	h := sha256.New()
	hState, err := GetHashState(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hState.EncodeToString(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeStringToState(t *testing.T) {
	h := sha256.New()
	if _, err := h.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}
	digest := h.Sum(nil)
	hState, err := GetHashState(h)
	if err != nil {
		t.Fatal(err)
	}
	stateString, err := hState.EncodeToString()
	if err != nil {
		t.Fatal(err)
	}

	otherH := sha256.New()
	otherState, err := DecodeStringToState(stateString)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetHashState(otherH, otherState); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(digest, otherH.Sum(nil)) {
		t.Fatalf("digest and otherH.Sum(nil) should be equal")
	}
}

func TestGetHashStateText(t *testing.T) {
	if _, err := GetHashStateText(sha256.New()); err != nil {
		t.Fatal(err)
	}
}

func TestNewHashWithState(t *testing.T) {
	helloHash := sha256.New()
	if _, err := helloHash.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}
	helloDigest := helloHash.Sum(nil)
	helloHashStateText, err := GetHashStateText(helloHash)
	if err != nil {
		t.Fatal(err)
	}

	helloHashCopy, err := NewHashWithStateText(helloHashStateText)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(helloDigest, helloHashCopy.Sum(nil)) {
		t.Fatalf("helloDigest should be equal to helloHashCopy.Sum(nil)")
	}
}

func TestContinuousHash(t *testing.T) {
	word1 := "hello"
	word2 := "world"

	completeHash := sha256.New()
	if _, err := completeHash.Write([]byte(word1)); err != nil {
		t.Fatal(err)
	}
	if _, err := completeHash.Write([]byte(word2)); err != nil {
		t.Fatal(err)
	}
	completeDigest := completeHash.Sum(nil)

	part1Hash := sha256.New()
	if _, err := part1Hash.Write([]byte(word1)); err != nil {
		t.Fatal(err)
	}
	part1HashText, err := GetHashStateText(part1Hash)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := NewHashWithStateText(part1HashText)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hash.Write([]byte(word2)); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(completeDigest, hash.Sum(nil)) {
		t.Fatalf("completeDigest should be equal to hash.Sum(nil)")
	}
}
