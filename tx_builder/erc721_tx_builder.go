package tx_builder

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/unifralabs/unifra-benchmark-tool/contract/erc721"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

var (
	Erc721Abi, _ = erc721.Erc721MetaData.GetAbi()
)

type ERC721TxBuilder struct {
	mnemonic        string
	url             string
	provider        *ethclient.Client
	rpcClient       *rpc_client.RpcClient
	gasEstimation   *big.Int
	gasPrice        *big.Int
	defaultValue    *big.Int
	nftName         string
	nftSymbol       string
	nftURL          string
	contractAddress *common.Address
	contract        *erc721.Erc721 // You'll need to generate this contract binding
	baseDeployer    *bind.TransactOpts
}

func NewERC721TxBuilder(mnemonic, url string) (*ERC721TxBuilder, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	return &ERC721TxBuilder{
		mnemonic:      mnemonic,
		url:           url,
		provider:      client,
		rpcClient:     rpcClient,
		gasEstimation: big.NewInt(0),
		gasPrice:      big.NewInt(0),
		defaultValue:  big.NewInt(0),
		nftName:       "ZEXTokens",
		nftSymbol:     "ZEXes",
		nftURL:        "https://really-valuable-nft-page.io",
	}, nil
}

func (e *ERC721TxBuilder) Initialize() error {
	_, privateKey, err := utils.DerivePrivateKeyFromMnemonic(e.mnemonic, 0)
	if err != nil {
		return fmt.Errorf("failed to derive private key: %v", err)
	}

	chainID, err := e.provider.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	baseGasPrice, err := e.rpcClient.GetGasPrice()
	if err != nil {
		return fmt.Errorf("failed to get gas price: %v", err)
	}

	e.baseDeployer, err = bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %v", err)
	}
	e.baseDeployer.GasPrice = baseGasPrice

	// Deploy the contract
	address, tx, instance, err := erc721.DeployErc721(e.baseDeployer, e.provider, e.nftName, e.nftSymbol)
	if err != nil {
		return fmt.Errorf("failed to deploy contract: %v", err)
	}

	_, err = bind.WaitDeployed(context.Background(), e.provider, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for contract deployment: %v", err)
	}

	e.contractAddress = &address
	e.contract = instance
	log.Info().Msgf("Contract deployed at address: %s", address.Hex())

	return nil
}

func ContructErc721Mint(tokenURI string) ([]byte, error) {
	input, err := Erc721Abi.Pack("createNFT", tokenURI)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (e *ERC721TxBuilder) EstimateGasForBaseTx() (*big.Int, error) {
	if e.contract == nil {
		return nil, fmt.Errorf("runtime not initialized")
	}

	input, err := ContructErc721Mint(e.nftURL)
	if err != nil {
		return nil, fmt.Errorf("failed to construct input: %v", err)
	}

	gasEstimation, err := e.rpcClient.EstimateGasForContractCall(e.baseDeployer.From, *e.contractAddress, input)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}
	log.Info().Msgf("ERC72 mint gasEstimation: %d", gasEstimation)

	e.gasEstimation = big.NewInt(int64(gasEstimation * 10))
	return e.gasEstimation, nil
}

func (e *ERC721TxBuilder) GetNFTSymbol() string {
	return e.nftSymbol
}

func (e *ERC721TxBuilder) GetValue() *big.Int {
	return e.defaultValue
}

func (e *ERC721TxBuilder) ConstructTransactions(accounts []*tooltypes.SenderAccount, numTx int) ([]*types.Transaction, error) {
	if e.contract == nil {
		return nil, fmt.Errorf("runtime not initialized")
	}

	chainID, err := e.provider.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	gasPrice, err := e.rpcClient.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	log.Info().Msgf("Chain ID: %s", chainID.String())
	log.Info().Msgf("Avg. gas price: %s", gasPrice.String())

	log.Info().Msgf("Constructing %s mint transactions...", e.nftName)
	bar := progressbar.Default(int64(numTx))

	transactions := make([]*types.Transaction, numTx)

	for i := 0; i < numTx; i++ {
		senderIndex := i % len(accounts)
		sender := accounts[senderIndex]

		input, err := ContructErc721Mint(e.nftURL)
		if err != nil {
			return nil, fmt.Errorf("failed to construct input: %v", err)
		}

		tx := types.NewTransaction(sender.GetNonce(), *e.contractAddress, new(big.Int), e.gasEstimation.Uint64(), gasPrice, input)

		transactions[i] = tx

		sender.IncrNonce()
		bar.Add(1)
	}

	log.Info().Msgf("Successfully constructed %d transactions", numTx)

	return transactions, nil
}
