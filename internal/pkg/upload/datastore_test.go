package upload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DataStore_All(t *testing.T) {
	store := DataStore{"data"}
	assert.Equal(t, "data/db.sqlite3", store.GetDBPath())
	assert.Equal(t, "data/blobs/testblob", store.GetBlobPath("testblob"))
	assert.Equal(t, "data/blobs/testblob.thumb.jpg", store.GetBlobThumbPath("testblob.jpg"))
}

func Test_DataStore_Thumbs(t *testing.T) {
	assert.Equal(t, "pic.thumb.jpg", GetThumbName("pic.jpg"))
	assert.Equal(t, "blob.thumb", GetThumbName("blob"))

	assert.Equal(t, "pic.jpg", RevertThumbName("pic.thumb.jpg"))
	assert.Equal(t, "blob", RevertThumbName("blob.thumb"))
}
