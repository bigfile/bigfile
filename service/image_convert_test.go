//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"image"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bigfile/bigfile/internal/util"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/gographics/gmagick"
	"github.com/stretchr/testify/assert"
)

func TestNewGm(t *testing.T) {
	gm := NewGm()
	imgName, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(imgName))
}

func TestGM_ImageCropValidate(t *testing.T) {
	gm := NewGm()
	imgName, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(imgName))
	assert.Nil(t, gm.ImageCrop(100, 100, 0, 0))
}

func TestGM_ImageThumb(t *testing.T) {
	gm := NewGm()
	imgName, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(imgName))
	assert.Nil(t, gm.ImageThumb(100, 100))
}

func TestGM_ImageZoom(t *testing.T) {
	gm := NewGm()
	imgName, down := models.NewImageForTest(t)
	defer func() {
		down(t)
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	assert.Nil(t, gm.MagickWand.ReadImage(imgName))
	assert.Nil(t, gm.ImageZoom(100, 100))
}

func TestImageConvert_Validate(t *testing.T) {

	var (
		ImageConvertSrv = &ImageConvert{
			Token: nil,
			File:  nil,
			IP:    nil,
		}

		err         error
		errValidate ValidateErrors
	)

	confirm := assert.New(t)
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	confirm.Nil(err)
	defer down(t)
	ImageConvertSrv.DB = trx

	errValidate = ImageConvertSrv.Validate()
	assert.NotNil(t, errValidate)
	confirm.NotNil(errValidate)
	confirm.True(errValidate.ContainsErrCode(10036))
	confirm.True(errValidate.ContainsErrCode(10037))
	confirm.True(errValidate.ContainsErrCode(10038))
	confirm.True(errValidate.ContainsErrCode(10039))
	confirm.True(errValidate.ContainsErrCode(10040))
	confirm.True(errValidate.ContainsErrCode(10036))
	confirm.True(errValidate.ContainsErrCode(10037))
	confirm.Contains(errValidate.Error(), "invalid token")
	confirm.Contains(errValidate.Error(), "invalid file")

	token.Path = "/test"
	confirm.Nil(trx.Save(token).Error)

	dir, err := models.CreateOrGetLastDirectory(&token.App, "/save/to", trx)
	confirm.Nil(err)

	ImageConvertSrv.Token = token
	ImageConvertSrv.File = dir

	errValidate = ImageConvertSrv.Validate()
	confirm.NotNil(errValidate)
	confirm.Contains(errValidate.Error(), "file can't be accessed by some tokens")
}

func TestImageConvert_Execute(t *testing.T) {
	tempDir := models.NewTempDirForTest()
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	imgName, downImg := models.NewImageForTest(t)
	img, err := ioutil.ReadFile(imgName)
	assert.Nil(t, err)
	defer func() {
		downImg(t)
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	assert.Nil(t, err)
	imgHash, err := util.Sha256Hash2String(img)
	assert.Nil(t, err)
	token.AvailableTimes = 1000
	assert.Nil(t, trx.Save(token).Error)
	file, err := models.CreateFileFromReader(&token.App, "/test/random.bytes", bytes.NewReader(img), int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, file.Object.Hash, imgHash)

	ImageConvertSrv := &ImageConvert{
		BaseService: BaseService{
			DB:       trx,
			RootPath: &tempDir,
		},
		Token:  token,
		File:   file,
		Type:   "thumb",
		Width:  100,
		Height: 200,
	}
	assert.Nil(t, ImageConvertSrv.Validate())
	//use thumb
	fileCreateValue, err := ImageConvertSrv.Execute(context.TODO())
	cc, _, err := image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 100)
	assert.Equal(t, cc.Height, 100)
	ImageConvertSrv.Width = 300
	ImageConvertSrv.Height = 200
	fileCreateValue, err = ImageConvertSrv.Execute(context.TODO())
	cc, _, err = image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 200)
	assert.Equal(t, cc.Height, 200)

	//use zoom
	ImageConvertSrv.Type = "zoom"
	ImageConvertSrv.Width = 300
	ImageConvertSrv.Height = 200
	fileCreateValue, err = ImageConvertSrv.Execute(context.TODO())
	cc, _, err = image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 300)
	assert.Equal(t, cc.Height, 200)

	ImageConvertSrv.Type = "zoom"
	ImageConvertSrv.Width = 100
	ImageConvertSrv.Height = 200
	fileCreateValue, err = ImageConvertSrv.Execute(context.TODO())
	cc, _, err = image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 100)
	assert.Equal(t, cc.Height, 200)

	//use crop
	ImageConvertSrv.Type = "crop"
	ImageConvertSrv.Width = 100
	ImageConvertSrv.Height = 200
	fileCreateValue, err = ImageConvertSrv.Execute(context.TODO())
	cc, _, err = image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 100)
	assert.Equal(t, cc.Height, 200)

	ImageConvertSrv.Type = "crop"
	ImageConvertSrv.Width = 300
	ImageConvertSrv.Height = 200
	fileCreateValue, err = ImageConvertSrv.Execute(context.TODO())
	cc, _, err = image.DecodeConfig(bytes.NewReader(fileCreateValue))
	assert.Nil(t, err)
	assert.Equal(t, cc.Width, 300)
	assert.Equal(t, cc.Height, 200)

}
