//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"github.com/gographics/gmagick"
)

type GM struct {
	MagickWand *gmagick.MagickWand
}

func NewGm() *GM {
	MagickWand := gmagick.NewMagickWand()
	gmagick.Initialize()
	return &GM{
		MagickWand: MagickWand,
	}
}

func (gm *GM) ImageThumb(width, height float64) error {
	w := gm.MagickWand.GetImageWidth()
	h := gm.MagickWand.GetImageHeight()
	x := float64(w) / float64(h)
	var targetW, targetH uint
	if height*x > width {
		targetW = uint(width)
		targetH = uint(width / x)
	} else {
		targetW = uint(height * x)
		targetH = uint(height)
	}
	err := gm.MagickWand.ResizeImage(targetW, targetH, gmagick.FILTER_LANCZOS, 1)
	if err != nil {
		return err
	}
	return nil
}

func (gm *GM) ImageCrop(width, height float64, left, top int) error {
	err := gm.MagickWand.CropImage(uint(width), uint(height), left, top)
	if err != nil {
		return err
	}
	return nil
}

func (gm *GM) ImageZoom(width, height float64) error {
	var left, top int
	var x, xW, xH float64

	w := gm.MagickWand.GetImageWidth()
	h := gm.MagickWand.GetImageHeight()
	xW = float64(w) / width
	xH = float64(h) / height

	if xW < xH {
		x = xW
	} else {
		x = xH
	}
	thumbWidth, thumbHeight := uint(float64(w)/x), uint(float64(h)/x)

	if err := gm.MagickWand.ResizeImage(thumbWidth, thumbHeight, gmagick.FILTER_LANCZOS, 1); err != nil {
		return err
	}

	top = int(thumbHeight-uint(height)) / 2
	left = int(thumbWidth-uint(width)) / 2

	if err := gm.MagickWand.CropImage(uint(width), uint(height), left, top); err != nil {
		return err
	}

	return nil
}
