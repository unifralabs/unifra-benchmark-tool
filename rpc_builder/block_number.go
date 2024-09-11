package rpc_builder

import (
	"slices"

	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func GenerateBlockNumbers(
	n int,
	startBlock int64,
	endBlock int64,
	sort bool,
	randomSeed *tooltypes.RandomSeed,
	network *string,
) ([]int64, error) {
	// Seed a generator
	rng, err := utils.GetRNG(randomSeed)
	if err != nil {
		return nil, err
	}

	// Generate blocks
	allBlocks := make([]int64, endBlock-startBlock+1)
	for i := range allBlocks {
		allBlocks[i] = startBlock + int64(i)
	}

	chosen := make([]int64, n)
	if n > len(allBlocks) {
		for i := range chosen {
			chosen[i] = allBlocks[rng.Intn(len(allBlocks))]
		}
	} else {
		perm := rng.Perm(len(allBlocks))
		for i := 0; i < n; i++ {
			chosen[i] = allBlocks[perm[i]]
		}
	}

	// Sort if required
	if sort {
		slices.Sort(chosen)
	}

	return chosen, nil
}
