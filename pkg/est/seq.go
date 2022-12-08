package est

import (
	"fmt"
	"image"

	"github.com/rs/zerolog/log"
)

type sequence struct {
	// Those slices always have the same length.
	// dx[i] is the assumed offset between frames[i] and frames[i+1].
	// scores[i] is the score of that assumed offset.
	dx     []int
	frames []image.Image
}

func processSequence(seq sequence) {
	dx, err := cleanupDx(seq.dx)
	if err != nil {
		log.Warn().Err(err).Msg("was not able to clean up sequence")
	}

	fmt.Println(dx) // TODO: assemble
}
