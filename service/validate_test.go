package service

import (
	"strings"
	"testing"

	"github.com/jinzhu/gorm"

	models "github.com/bigfile/bigfile/databases/mdoels"

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
	assert.True(t, ValidatePath("/test"))
	assert.True(t, ValidatePath("test"))
	assert.True(t, ValidatePath("/test/"))
	assert.True(t, ValidatePath("/test/hello"))
	assert.False(t, ValidatePath("/test//"))
	name := strings.Repeat("s", 255)
	assert.True(t, ValidatePath("/test/"+name+"/"))
	assert.False(t, ValidatePath("/test/"+name+"1222/"))
}
