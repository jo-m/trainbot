// Package upload deals with uploading images and database to a FTP server.
package upload

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/db"
	"github.com/rs/zerolog/log"
)

const (
	dbFile    = "db.sqlite3"
	dbBakFile = "db.sqlite3.bak"
)

// Uploader uploads files to a remote location.
type Uploader interface {
	Upload(ctx context.Context, remotePath string, contents io.Reader) error
	AtomicUpload(ctx context.Context, remotePath string, contents io.Reader) error
	Close() error
}

func serverBlobPath(blobName string) string {
	return path.Join("blobs", blobName)
}

func uploadFile(ctx context.Context, uploader Uploader, localPath, remotePath string, atomic bool) error {
	log.Info().Str("local", localPath).Str("remote", remotePath).Msg("uploading file")
	// #nosec G304
	f, err := os.Open(localPath)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer f.Close()

	if atomic {
		return uploader.AtomicUpload(ctx, remotePath, f)
	}

	return uploader.Upload(ctx, remotePath, f)
}

// All uploads all pending trains, until an error is hit or there are no more pending uploads.
// Also updates the database, and uploads the updated database.
func All(ctx context.Context, dbx *sqlx.DB, uploader Uploader, dataDir, blobsDir string) (int, error) {
	var nUploads int
	for {
		toUpload, err := db.GetNextUpload(dbx)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Debug().Msg("no more files to upload")
				break
			}

			log.Err(err).Send()
			return 0, err
		}

		log.Info().Str("img", toUpload.ImgPath).Str("gif", toUpload.GIFPath).Int64("id", toUpload.ID).Msg("uploading")

		err = uploadFile(ctx, uploader, filepath.Join(blobsDir, toUpload.ImgPath), serverBlobPath(toUpload.ImgPath), false)
		if err != nil {
			log.Err(err).Send()
			return 0, err
		}

		err = uploadFile(ctx, uploader, filepath.Join(blobsDir, toUpload.GIFPath), serverBlobPath(toUpload.GIFPath), false)
		if err != nil {
			log.Err(err).Send()
			return 0, err
		}

		err = db.SetUploaded(dbx, toUpload.ID)
		if err != nil {
			log.Err(err).Send()
			return 0, err
		}

		nUploads++
	}

	// Do not upload db if no files were uploaded.
	if nUploads == 0 {
		return nUploads, nil
	}

	// Create db backup.
	log.Info().Msg("creating db backup")
	dbBakPath := filepath.Join(dataDir, dbBakFile)
	err := db.Backup(dbx, dbBakPath)
	if err != nil {
		log.Err(err).Send()
		return 0, err
	}

	return nUploads, uploadFile(ctx, uploader, dbBakPath, dbBakFile, true)
}
