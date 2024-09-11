package tx_builder

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

type EOATxBuilder struct {
	mnemonic  string
	url       string
	provider  *ethclient.Client
	rpcClient *rpc_client.RpcClient

	gasEstimation *big.Int
	gasPrice      *big.Int

	defaultValue *big.Int
}

func NewEOATxBuilder(mnemonic, url string) (*EOATxBuilder, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	return &EOATxBuilder{
		mnemonic:      mnemonic,
		url:           url,
		provider:      client,
		rpcClient:     rpcClient,
		gasEstimation: big.NewInt(0),
		gasPrice:      big.NewInt(0),
		defaultValue:  big.NewInt(1e14), // 0.0001 ETH
	}, nil
}

func (e *EOATxBuilder) Initialize() error {
	return nil
}

func (e *EOATxBuilder) EstimateGasForBaseTx() (*big.Int, error) {
	// Create a wallet from the mnemonic
	wallet, err := hdwallet.NewFromMnemonic(e.mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Derive the 'from' address (index 0)
	fromPath := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	fromAccount, err := wallet.Derive(fromPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to derive 'from' address: %w", err)
	}
	from := fromAccount.Address

	// Derive the 'to' address (index 1)
	toPath := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1")
	toAccount, err := wallet.Derive(toPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to derive 'to' address: %w", err)
	}
	to := toAccount.Address

	// Estimate gas for a simple value transfer
	gasLimit, err := e.provider.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: e.defaultValue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	log.Info().Msgf("gasEstimation: %d", gasLimit)

	e.gasEstimation = new(big.Int).SetUint64(gasLimit)
	return e.gasEstimation, nil
}

func (e *EOATxBuilder) GetValue() *big.Int {
	return e.defaultValue
}

func (e *EOATxBuilder) ConstructTransactions(accounts []*tooltypes.SenderAccount, numTx int) ([]*types.Transaction, error) {

	chainID, err := e.provider.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	gasPrice, err := e.rpcClient.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	log.Info().Msgf("Chain ID: %s", chainID.String())
	log.Info().Msgf("Avg. gas price: %s", gasPrice.String())

	log.Info().Msg("Constructing value transfer transactions...")
	bar := progressbar.Default(int64(numTx))

	transactions := make([]*types.Transaction, numTx)

	for i := 0; i < numTx; i++ {
		senderIndex := i % len(accounts)
		receiverIndex := (i + 1) % len(accounts)

		sender := accounts[senderIndex]
		receiver := accounts[receiverIndex]

		tx := types.NewTransaction(sender.GetNonce(), receiver.GetAddress(), e.defaultValue, e.gasEstimation.Uint64(), gasPrice, nil)

		transactions[i] = tx

		sender.IncrNonce()
		bar.Add(1)
	}

	log.Info().Msgf("Successfully constructed %d transactions", numTx)

	return transactions, nil
}
