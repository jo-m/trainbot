// Package upload deals with uploading images and database to a FTP server.
package upload

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/db"
	"github.com/rs/zerolog/log"
)

const (
	dbBakFile = "db.sqlite3.bak"
)

// Uploader is an interface for interaction with a remote file storage location.
type Uploader interface {
	// Upload uploads a file.
	Upload(ctx context.Context, remotePath string, contents io.Reader) error
	// AtomicUpload uploads a file, trying to swap out the file in an atomic operation.
	AtomicUpload(ctx context.Context, remotePath string, contents io.Reader) error
	// ListFiles lists all regular files in a remote directory.
	// Any non-regular files (e.g. directories) are to be ignored.
	ListFiles(ctx context.Context, remotePath string) ([]string, error)
	// DeleteFile deletes a regular file at the given remote path.
	DeleteFile(ctx context.Context, remotePath string) error
	// Close terminates the connection and frees any resources.
	Close() error
}

func serverBlobPath(blobName string) string {
	return path.Join(blobsDir, blobName)
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
func All(ctx context.Context, store DataStore, dbx *sqlx.DB, uploader Uploader) (int, error) {
	var nUploads int
	for {
		toUpload, err := db.GetNextUpload(dbx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Debug().Msg("no more files to upload")
				break
			}

			log.Err(err).Send()
			return 0, err
		}

		log.Info().Str("img", toUpload.ImgPath).Str("gif", toUpload.GIFPath).Int64("id", toUpload.ID).Msg("uploading")

		err = uploadFile(ctx, uploader, store.GetBlobPath(toUpload.ImgPath), serverBlobPath(toUpload.ImgPath), false)
		if err != nil {
			log.Err(err).Send()
			if !errors.Is(err, fs.ErrNotExist) {
				return 0, err
			}
		}

		err = uploadFile(ctx, uploader, store.GetBlobThumbPath(toUpload.ImgPath), serverBlobPath(GetThumbName(toUpload.ImgPath)), false)
		if err != nil {
			log.Err(err).Send()
			if !errors.Is(err, fs.ErrNotExist) {
				return 0, err
			}
		}

		err = uploadFile(ctx, uploader, store.GetBlobPath(toUpload.GIFPath), serverBlobPath(toUpload.GIFPath), false)
		if err != nil {
			log.Err(err).Send()
			if !errors.Is(err, fs.ErrNotExist) {
				return 0, err
			}
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
	err := db.Backup(dbx, store.GetDataPath(dbBakFile))
	if err != nil {
		log.Err(err).Send()
		return 0, err
	}

	return nUploads, uploadFile(ctx, uploader, store.GetDataPath(dbBakFile), dbFile, true)
}

// CleanupOrphanedRemoteBlobs removes from the remote storage all blobs which are unknown to the database.
func CleanupOrphanedRemoteBlobs(ctx context.Context, dbx *sqlx.DB, uploader Uploader) (int, error) {
	// Get list of blobs from remote.
	remoteBlobs, err := uploader.ListFiles(ctx, blobsDir)
	if err != nil {
		return 0, err
	}
	sort.Strings(remoteBlobs)

	// Map of blobs existing in the database, for comparison.
	knownBlobs, err := db.GetAllBlobs(dbx)
	if err != nil {
		return 0, err
	}

	var nDeletions int
	for _, remoteBlob := range remoteBlobs {
		_, known := knownBlobs[remoteBlob]
		_, knownThumb := knownBlobs[RevertThumbName(remoteBlob)]
		if !known && !knownThumb {
			log.Info().Str("remoteBlob", remoteBlob).Msg("orphaned blob, deleting")
			err := uploader.DeleteFile(ctx, serverBlobPath(remoteBlob))
			if err != nil {
				log.Err(err).Send()
				return 0, err
			}
			nDeletions++
		}
	}

	return nDeletions, nil
}
