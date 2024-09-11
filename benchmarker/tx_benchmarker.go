package benchmarker

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/unifralabs/unifra-benchmark-tool/distributor"
	"github.com/unifralabs/unifra-benchmark-tool/outputter"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	"github.com/unifralabs/unifra-benchmark-tool/stats"
	"github.com/unifralabs/unifra-benchmark-tool/tx_builder"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

type TxBenchmarker struct {
	txType           tooltypes.TxType
	mnemonic         string
	url              string
	provider         *ethclient.Client
	rpcClient        *rpc_client.RpcClient
	txBuilder        tooltypes.TxBuilder
	subAccountsCount int
	transactionCount int
	batchSize        int
	outputDir        string
	accountIndexes   []int
}

func NewTxBenchmarker(client *ethclient.Client, mnemonic, url string, txType tooltypes.TxType,
	subAccountsCount int, transactionCount int, batchSize int, outputDir string) (*TxBenchmarker, error) {

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	var txBuilder tooltypes.TxBuilder
	switch txType {
	case tooltypes.EOA:
		txBuilder, err = tx_builder.NewEOATxBuilder(mnemonic, url)
	case tooltypes.ERC20:
		txBuilder, err = tx_builder.NewERC20TxBuilder(mnemonic, url)
	case tooltypes.ERC721:
		txBuilder, err = tx_builder.NewERC721TxBuilder(mnemonic, url)
	default:
		return nil, fmt.Errorf("unknown runtime mode: %s", txType)
	}

	return &TxBenchmarker{
		txType:           txType,
		mnemonic:         mnemonic,
		url:              url,
		provider:         client,
		rpcClient:        rpcClient,
		txBuilder:        txBuilder,
		subAccountsCount: subAccountsCount,
		transactionCount: transactionCount,
		batchSize:        batchSize,
		outputDir:        outputDir,
	}, nil
}

func (t *TxBenchmarker) Initialize() error {

	err := t.txBuilder.Initialize()
	if err != nil {
		return err
	}

	// Distribute the native currency funds
	d, err := distributor.NewDistributor(t.mnemonic, t.subAccountsCount, t.transactionCount, t.txBuilder, t.url)
	if err != nil {
		return err
	}

	accountIndexes, err := d.Distribute()
	if err != nil {
		return err
	}
	t.accountIndexes = accountIndexes

	// Distribute the token funds, if any
	if t.txType == tooltypes.ERC20 {
		td, err := distributor.NewTokenDistributor(t.mnemonic, accountIndexes, t.transactionCount, t.txBuilder.(tooltypes.Erc20TxBuilder))
		if err != nil {
			return err
		}

		if _, err := td.DistributeTokens(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TxBenchmarker) Run() error {
	ctx := NewTxBenchmarkerContext(t.accountIndexes, t.transactionCount, t.batchSize, t.mnemonic, t.url)
	txHashes, err := BuildAndSendTransactions(t.provider, t.txBuilder, ctx)
	if err != nil {
		return err
	}

	// Collect the data
	collectorData, err := stats.GenerateStats(t.provider, txHashes, t.batchSize)
	if err != nil {
		return err
	}

	// Output the data if needed
	if t.outputDir != "" {
		err := outputter.OutputData(collectorData, t.outputDir)
		return err
	}

	return nil
}

type TxBenchmarkerContext struct {
	AccountIndexes []int
	NumTxs         int
	BatchSize      int
	Mnemonic       string
	URL            string
}

func NewTxBenchmarkerContext(accountIndexes []int, numTxs, batchSize int, mnemonic, url string) *TxBenchmarkerContext {
	return &TxBenchmarkerContext{
		AccountIndexes: accountIndexes,
		NumTxs:         numTxs,
		BatchSize:      batchSize,
		Mnemonic:       mnemonic,
		URL:            url,
	}
}

func BuildAndSendTransactions(ethclient *ethclient.Client, runtime tooltypes.TxBuilder, ctx *TxBenchmarkerContext) ([]*types.Transaction, error) {

	// Get the account metadata
	accounts, err := utils.GetSenderAccounts(ethclient, ctx.Mnemonic, ctx.AccountIndexes, ctx.NumTxs)
	if err != nil {
		return nil, err
	}

	// Construct the transactions
	rawTransactions, err := runtime.ConstructTransactions(accounts, ctx.NumTxs)
	if err != nil {
		return nil, err
	}

	// Sign the transactions
	signedTransactions, err := utils.SignTransactions(ethclient, accounts, rawTransactions)
	if err != nil {
		return nil, err
	}

	batches := utils.GenerateBatches(signedTransactions, ctx.BatchSize)

	// Send the transactions in batches
	_, err = utils.BatchSendRawTransactions(batches, ctx.URL)
	if err != nil {
		return nil, err
	}
	return signedTransactions, nil
}
