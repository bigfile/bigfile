//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"testing"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/gographics/gmagick"
	"github.com/stretchr/testify/assert"
)

func TestGM_ImageCropValidate(t *testing.T) {
	gm := NewGm()
	f, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(f.Name()))
	assert.Nil(t, gm.ImageCrop(100, 100, 0, 0))
}

func TestGM_ImageThumb(t *testing.T) {
	gm := NewGm()
	f, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(f.Name()))
	assert.Nil(t, gm.ImageThumb(100, 100))
}

func TestGM_ImageZoom(t *testing.T) {
	gm := NewGm()
	f, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(f.Name()))
	assert.Nil(t, gm.ImageZoom(100, 100))
}
