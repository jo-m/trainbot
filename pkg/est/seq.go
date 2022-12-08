package est

import (
	"errors"
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

// allowed to remove trailing values from dx, but not values from the beginnign
func cleanupDx(dx []int) ([]int, error) {
	if len(dx) < 10 {
		return nil, errors.New("len(x) must be >= 10")
	}
	if dx[0] == 0 {
		return nil, errors.New("first dx value cannot be 0")
	}

	// remove trailing zeros
	for dx[len(dx)-1] == 0 {
		dx = dx[:len(dx)-1]
	}

	dxFit, err := fitDx(dx)
	if err != nil {
		return nil, err
	}

	return dxFit, nil
}

func processSequence(seq sequence) error {
	if len(seq.dx) != len(seq.frames) {
		log.Panic().Msg("length of frames and dx should be equal, this should not happen")
	}

	// remove trailing zeros
	for seq.dx[len(seq.dx)-1] == 0 {
		seq.dx = seq.dx[:len(seq.dx)-1]
		seq.frames = seq.frames[:len(seq.frames)-1]
	}

	// checks
	if len(seq.dx) < 10 {
		return errors.New("sequence too short")
	}
	if seq.dx[0] == 0 {
		return errors.New("seq.dx cannot start with a zero")
	}

	var err error
	seq.dx, err = fitDx(seq.dx)
	if err != nil {
		return errors.New("was not able to fit the sequence")
	}

	fmt.Println(seq.dx) // TODO: assemble

	return nil
}
