package rpc_builder

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/unifralabs/unifra-benchmark-tool/types"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

// GenerateTestEthGetBalance generates a sequence of VegetaAttacks for testing eth_getBalance
func GenerateTestEthGetBalance(params tooltypes.TestGenerationParameters) ([]tooltypes.VegetaAttack, error) {
	nCalls, err := tooltypes.EstimateCallCount(params.Rates, params.Durations, nil)
	if err != nil {
		return nil, err
	}

	calls, err := GenerateCallsEthGetEthBalance(nCalls, &params.Network, nil, nil, &params.RandomSeed)
	if err != nil {
		return nil, err
	}

	return tooltypes.CreateLoadTest(calls,
		params.Rates, params.Durations, params.VegetaArgs, true)
}

func GenerateCallsEthGetEthBalance(
	nCalls int,
	network *string,
	addresses []string,
	blockNumbers []int64,
	randomSeed *tooltypes.RandomSeed,
) ([]*types.JsonrpcMessage, error) {
	if blockNumbers == nil {
		var err error
		blockNumbers, err = GenerateBlockNumbers(
			nCalls,
			10_000_000,
			16_000_000,
			true,
			randomSeed,
			network,
		)
		if err != nil {
			return nil, err
		}
	}
	if addresses == nil {
		var err error
		addresses, err = GenerateContractAddresses(
			nCalls,
			network,
			randomSeed,
		)
		if err != nil {
			return nil, err
		}
	}

	calls := make([]*types.JsonrpcMessage, len(addresses))
	for i := range addresses {
		blockNumber := rpc.BlockNumber(blockNumbers[i])
		address := common.HexToAddress(addresses[i])
		calls[i] = ConstructEthGetBalance(address, &blockNumber)
	}
	return calls, nil
}

// TestGenerationParams represents the parameters for test generation
type TestGenerationParams struct {
	Rates      []int
	Duration   *int
	Durations  []int
	Network    string
	VegetaArgs interface{}
	RandomSeed *tooltypes.RandomSeed
}

// GenerateCallsParams represents the parameters for generating calls
type GenerateCallsParams struct {
	NCalls     int
	Network    string
	RandomSeed *tooltypes.RandomSeed
}
