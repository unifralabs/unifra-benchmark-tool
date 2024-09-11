package utils

import (
	"time"

	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"golang.org/x/exp/rand"
)

func GetRNG(randomSeed *tooltypes.RandomSeed) (*rand.Rand, error) {
	var seed int64

	if randomSeed == nil {
		seed = time.Now().UnixNano()
	} else {
		seed = int64(*randomSeed)
	}

	source := rand.NewSource(uint64(seed))
	return rand.New(source), nil
}
