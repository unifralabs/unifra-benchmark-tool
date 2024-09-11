package rpc_builder

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func ConstructEthGetBlockByNumber(blockNumber *rpc.BlockNumber, includeFullTransactions bool) *types.JsonrpcMessage {
	encodedBlockNumber := encodeBlockNumber(*blockNumber)

	parameters := []interface{}{encodedBlockNumber, includeFullTransactions}
	return utils.NewJsonrpcMessage("eth_getBlockByNumber", parameters)
}

func ConstructEthGetBalance(address common.Address, blockNumber *rpc.BlockNumber) *types.JsonrpcMessage {
	if blockNumber == nil {
		latest := rpc.LatestBlockNumber
		blockNumber = &latest
	}
	encodedBlockNumber := encodeBlockNumber(*blockNumber)
	return utils.NewJsonrpcMessage("eth_getBalance", []interface{}{address, encodedBlockNumber})
}

func encodeBlockNumber(blockNumber rpc.BlockNumber) string {
	if blockNumber == rpc.LatestBlockNumber {
		return "latest"
	}
	return blockNumber.String()
}
