package service

import (
	"strings"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestValidateApp(t *testing.T) {
	var (
		app     *models.App
		err     error
		conn    *gorm.DB
		confirm = assert.New(t)
		down    func(*testing.T)
	)

	app, conn, down, err = models.NewAppForTest(nil, t)
	confirm.Nil(err)
	defer down(t)

	err = ValidateApp(conn, nil)
	confirm.NotNil(err)
	confirm.Contains(err.Error(), "invalid application")

	err = ValidateApp(conn, &models.App{UID: "fake uid"})
	confirm.NotNil(err)
	confirm.Contains(err.Error(), "record not found")

	err = ValidateApp(conn, app)
	confirm.Nil(err)
}

func TestValidatePath(t *testing.T) {
	assert.True(t, ValidatePath("/"))
	assert.True(t, ValidatePath("/test"))
	assert.True(t, ValidatePath("test"))
	assert.True(t, ValidatePath("/test/"))
	assert.True(t, ValidatePath("/test/hello"))
	assert.False(t, ValidatePath("/test//"))
	name := strings.Repeat("s", 255)
	assert.True(t, ValidatePath("/test/"+name+"/"))
	assert.False(t, ValidatePath("/test/"+name+"1222/"))
}

func TestValidateToken(t *testing.T) {
	var (
		err   error
		trx   *gorm.DB
		token *models.Token
		down  func(t2 *testing.T)
	)

	token, trx, down, err = models.NewTokenForTest(nil, t, "/test/to", nil, nil, nil, 1000, int8(0))
	assert.Nil(t, err)
	defer down(t)

	err = ValidateToken(trx, nil, false, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	err = ValidateToken(trx, nil, false, &models.Token{UID: bson.NewObjectId().Hex()})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record not found")

	assert.Nil(t, ValidateToken(trx, nil, false, token))
}
