package upload

import (
	"path/filepath"
	"strings"
)

const (
	// Relative to data dir.
	dbFile = "db.sqlite3"
	// Relative to data dir.
	blobsDir = "blobs"
)

// DataStore is a utility to centralize data store file system paths and access.
type DataStore struct {
	DataDir string `arg:"--data-dir,env:DATA_DIR" help:"Directory to store output data" default:"data" placeholder:"DIR"`
}

// GetDataPath gets the path to a file in the top level data directory.
func (d DataStore) GetDataPath(fpath string) string {
	return filepath.Join(d.DataDir, fpath)
}

// GetDBPath gets the path to the database file.
func (d DataStore) GetDBPath() string {
	return d.GetDataPath(dbFile)
}

// GetBlobPath gets the path to a blob.
func (d DataStore) GetBlobPath(blobName string) string {
	return filepath.Join(d.DataDir, blobsDir, blobName)
}

// GetThumbName gets the file name of a blob thumbnail.
func GetThumbName(blobName string) string {
	ext := filepath.Ext(blobName)
	name := strings.TrimSuffix(blobName, ext)

	return name + ".thumb" + ext
}

// RevertThumbName inverts the result of GetThumbName(), i.e. converts a thumbnail
// name back to the initial file name.
func RevertThumbName(thumbName string) string {
	ext := filepath.Ext(thumbName)
	withThumbExt := strings.TrimSuffix(thumbName, ext)

	thumbExt := filepath.Ext(withThumbExt)
	name := strings.TrimSuffix(withThumbExt, thumbExt)

	if thumbExt == "" {
		return name
	}

	return name + ext
}

// GetBlobThumbPath gets the path to a blob thumbnail.
func (d DataStore) GetBlobThumbPath(blobName string) string {
	return d.GetBlobPath(GetThumbName(blobName))
}
