package utils

import (
	"github.com/unifralabs/unifra-benchmark-tool/types"
	"golang.org/x/exp/rand"
)

func NewJsonrpcMessage(method string, parameters []interface{}) *types.JsonrpcMessage {
	return &types.JsonrpcMessage{
		Version: "2.0",
		Method:  method,
		Params:  parameters,
		ID:      rand.Int63n(1e18) + 1, // Generate a random ID between 1 and 1e18
	}
}
