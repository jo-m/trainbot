/*
Small helper binary to find orphaned blobs so they can be cleaned up.
Usage:

	cd data/blobs
	find . > blobs.txt
	go run ./cmd/cleanup/ > missing.txt
	# Now, manually inspect missing.txt.
	cat missing.txt
	# And run it if OK.
	source missing.txt
*/
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"jo-m.ch/go/trainbot/internal/pkg/db"
	"jo-m.ch/go/trainbot/internal/pkg/logging"
	"jo-m.ch/go/trainbot/internal/pkg/upload"
)

type config struct {
	logging.LogConfig

	upload.DataStore
}

func (c *config) mustOpenDB() *sqlx.DB {
	dbx, err := db.Open(c.GetDBPath())
	if err != nil {
		log.Panic().Err(err).Msg("could not create/open database")
	}

	return dbx
}

func parseCheckArgs() config {
	c := config{}
	arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	return c
}

// Load a file list which was generated using
//
//	cd data/blobs
//	find . > blobs.txt
func loadFilesList(name string) []string {
	// #nosec G304
	f, err := os.Open(name)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	ret := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		ret = append(ret, line[2:])
	}

	return ret
}

func main() {
	c := parseCheckArgs()

	dbx := c.mustOpenDB()
	defer dbx.Close()

	dbBlobs, err := db.GetAllBlobs(dbx)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// Add thumbs.
	for k := range dbBlobs {
		dbBlobs[upload.GetThumbName(k)] = struct{}{}
	}

	// Load on disk file names.
	files := loadFilesList("blobs.txt")

	// Check.
	missing := 0
	for _, file := range files {
		_, inDB := dbBlobs[file]
		if !inDB {
			fmt.Printf("rm -f %s\n", file)
			missing++
		}
	}

	fmt.Println("# missing: ", missing)
}
