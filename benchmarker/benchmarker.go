package benchmarker

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/config"
	"github.com/unifralabs/unifra-benchmark-tool/db"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

type Benchmarker struct {
	cfg              *config.EnvConfig
	client           *ethclient.Client
	rpcClient        *rpc_client.RpcClient
	dbClient         *db.Client
	node             tooltypes.Node
	nodes            tooltypes.Nodes
	eoaTxBenchmarker *TxBenchmarker
	rpcBenchmarker   *RpcBenchmarker
}

func NewBenchmarker(cfg *config.EnvConfig) (*Benchmarker, error) {
	client, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	dbClient, err := db.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating db client: %s", err)
	}

	nodeStr := cfg.NodeName + "=" + cfg.RpcUrl
	node, err := utils.ParseNode(nodeStr, true)
	if err != nil {
		return nil, fmt.Errorf("error parsing node: %s", err)
	}
	nodes := tooltypes.Nodes{node.Name: node}
	// log.Info().Msgf("node: %s", node)

	eoaTxBenchmarker, err := NewTxBenchmarker(client, cfg.AdminAccountMnemonic, cfg.RpcUrl, tooltypes.EOA, cfg.NumTestAccounts, 60,
		cfg.SendTransactionBatchSize, cfg.OutputDir)
	if err != nil {
		return nil, err
	}

	rpcBenchmarker, err := NewRpcBenchmarker(cfg, nodes)
	if err != nil {
		return nil, err
	}

	return &Benchmarker{
		cfg:              cfg,
		client:           client,
		rpcClient:        rpcClient,
		dbClient:         dbClient,
		node:             node,
		nodes:            nodes,
		eoaTxBenchmarker: eoaTxBenchmarker,
		rpcBenchmarker:   rpcBenchmarker,
	}, nil
}

func (b *Benchmarker) Initialize() error {
	err := b.eoaTxBenchmarker.Initialize()
	if err != nil {
		return err
	}
	return nil
}

func (b *Benchmarker) RunBenchmarks(ctx context.Context) {

	err := b.eoaTxBenchmarker.Run()
	if err != nil {
		log.Error().Msgf("Error occurred when running EOA benchmarker: %v", err)
	}

	err = b.rpcBenchmarker.Run()
	if err != nil {
		log.Error().Msgf("Error occurred when running RPC benchmarker: %v", err)
	}
}
