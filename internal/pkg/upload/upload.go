// Package upload deals with uploading images and database to a FTP server.
package upload

import (
	"context"
	"database/sql"
	"fmt"
	"net/textproto"
	"os"
	"path"

	"github.com/jlaffaye/ftp"
	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/db"
	"github.com/rs/zerolog/log"
)

const (
	dbFile    = "db.sqlite3"
	dbBakFile = "db.sqlite3.bak"
)

// FTPConfig is the configuration to connect to a FTP server.
type FTPConfig struct {
	Host string
	Port uint16
	// User is expected to have the data dir as root/home directory.
	User, Pass string
}

func serverBlobPath(blobName string) string {
	return "blobs/" + blobName
}

func isFTPErr(err error, code int) bool {
	if errF, ok := err.(*textproto.Error); ok {
		return errF.Code == code
	}
	return false
}

func createDirs(conn *ftp.ServerConn) error {
	err := conn.MakeDir("blobs")
	if isFTPErr(err, 550) {
		return nil
	}
	log.Err(err).Send()
	return err
}

func upload(conn *ftp.ServerConn, localPath, remotePath string) error {
	log.Info().Str("local", localPath).Str("remote", remotePath).Msg("uploading file")
	f, err := os.Open(localPath)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer f.Close()

	return conn.Stor(remotePath, f)
}

// Run uploads all pending blobs to remote storage and updates the database.
// Will also create and upload a backup of the database if necessary.
func Run(ctx context.Context, dbx *sqlx.DB, ftpConf FTPConfig, dataDir, blobsDir string) error {
	// FTP setup.
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", ftpConf.Host, ftpConf.Port), ftp.DialWithContext(ctx))
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer conn.Quit()

	err = conn.Login(ftpConf.User, ftpConf.Pass)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	log.Info().Msg("connected to FTP server")

	err = createDirs(conn)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	// Upload each train.
	var nUploads int
	for {
		toUpload, err := db.GetNextUpload(dbx)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Debug().Msg("no more files to upload")
				break
			}

			log.Err(err).Send()
			return err
		}

		log.Info().Str("img", toUpload.ImgPath).Str("gif", toUpload.GIFPath).Int64("id", toUpload.ID).Msg("uploading")

		err = upload(conn, path.Join(blobsDir, toUpload.ImgPath), serverBlobPath(toUpload.ImgPath))
		if err != nil {
			log.Err(err).Send()
			return err
		}

		err = upload(conn, path.Join(blobsDir, toUpload.GIFPath), serverBlobPath(toUpload.GIFPath))
		if err != nil {
			log.Err(err).Send()
			return err
		}

		err = db.SetUploaded(dbx, toUpload.ID)
		if err != nil {
			log.Err(err).Send()
			return err
		}

		nUploads++
	}

	// Do not upload db if no files were uploaded.
	if nUploads == 0 {
		return nil
	}

	// Create db backup.
	log.Info().Msg("creating db backup")
	dbBakPath := path.Join(dataDir, dbBakFile)
	err = db.Backup(dbx, dbBakPath)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	// Upload it.
	err = upload(conn, dbBakPath, dbBakFile)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	// And rename over old file.
	return conn.Rename(dbBakFile, dbFile)
}
