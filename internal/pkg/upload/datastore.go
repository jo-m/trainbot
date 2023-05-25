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
func (d DataStore) GetBlobPath(blobname string) string {
	return filepath.Join(d.DataDir, blobsDir, blobname)
}

// GetThumbName gets the file name of a blob thumbnail.
func GetThumbName(blobname string) string {
	ext := filepath.Ext(blobname)
	name := strings.TrimSuffix(blobname, ext)
	return name + ".thumb" + ext
}

// GetBlobThumbPath gets the path to a blob thumbnail.
func (d DataStore) GetBlobThumbPath(blobname string) string {
	return d.GetBlobPath(GetThumbName(blobname))
}
