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
	"github.com/unifralabs/unifra-benchmark-tool/contract/erc20"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

var (
	Erc20Abi, _ = erc20.Erc20MetaData.GetAbi()
)

type ERC20TxBuilder struct {
	mnemonic      string
	url           string
	provider      *ethclient.Client
	rpcClient     *rpc_client.RpcClient
	gasEstimation *big.Int
	gasPrice      *big.Int
	defaultValue  *big.Int

	defaultTransferValue *big.Int
	totalSupply          *big.Int
	coinName             string
	coinSymbol           string
	contractAddress      *common.Address
	contract             *erc20.Erc20
	baseDeployer         *bind.TransactOpts
}

func NewERC20TxBuilder(mnemonic, url string) (*ERC20TxBuilder, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	return &ERC20TxBuilder{
		mnemonic:             mnemonic,
		url:                  url,
		provider:             client,
		rpcClient:            rpcClient,
		gasEstimation:        big.NewInt(0),
		gasPrice:             big.NewInt(0),
		defaultValue:         big.NewInt(0),
		defaultTransferValue: big.NewInt(1),
		totalSupply:          big.NewInt(500000000000),
		coinName:             "Zex Coin",
		coinSymbol:           "ZEX",
	}, nil
}

func (e *ERC20TxBuilder) Initialize() error {
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
	e.baseDeployer.GasLimit = 2000000

	// Deploy the contract
	address, tx, instance, err := erc20.DeployErc20(e.baseDeployer, e.provider, e.totalSupply, e.coinName, e.coinSymbol)
	if err != nil {
		return fmt.Errorf("failed to deploy contract: %v", err)
	}

	_, err = bind.WaitDeployed(context.Background(), e.provider, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for contract deployment: %v", err)
	}

	receipt, err := e.provider.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to wait for contract deployment: %v", err)
	}
	status := receipt.Status
	fmt.Printf("Receipt status: %d\n", status)
	if status != 1 {
		log.Error().Msgf("Deploy deployment failed for contract: %s", address.Hex())
		return fmt.Errorf("failed to deploy contract: %v", err)
	}

	e.contractAddress = &address
	e.contract = instance
	log.Info().Msgf("Contract deployed at address: %s", address.Hex())

	balance, err := e.GetTokenBalance(e.baseDeployer.From)
	if err != nil {
		return fmt.Errorf("failed to get token balance: %v", err)
	}

	log.Info().Msgf("Balance of deployer %s: %s", e.baseDeployer.From.Hex(), balance)

	return nil
}

func ContructErc20Transfer(receiver common.Address, numTokens *big.Int) ([]byte, error) {
	input, err := Erc20Abi.Pack("transfer", receiver, numTokens)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (e *ERC20TxBuilder) EstimateGasForBaseTx() (*big.Int, error) {
	if e.contract == nil {
		return nil, fmt.Errorf("runtime not initialized")
	}

	toAddress, err := utils.DeriveAddressFromMnemonic(e.mnemonic, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to derive toAddress: %v", err)
	}

	input, err := ContructErc20Transfer(*toAddress, e.defaultTransferValue)
	if err != nil {
		return nil, fmt.Errorf("failed to construct input: %v", err)
	}

	gasEstimation, err := e.rpcClient.EstimateGasForContractCall(e.baseDeployer.From, *e.contractAddress, input)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	e.gasEstimation = big.NewInt(int64(gasEstimation))
	return e.gasEstimation, nil
}

func (e *ERC20TxBuilder) GetTransferValue() (*big.Int, error) {
	return e.defaultTransferValue, nil
}

func (e *ERC20TxBuilder) GetTokenBalance(address common.Address) (*big.Int, error) {
	if e.contract == nil {
		return nil, fmt.Errorf("runtime not initialized")
	}

	balance, err := e.contract.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get token balance: %v", err)
	}

	return balance, nil
}

func (e *ERC20TxBuilder) GetSupplierBalance() (*big.Int, error) {
	return e.GetTokenBalance(e.baseDeployer.From)
}

func (e *ERC20TxBuilder) FundAccount(to common.Address, amount *big.Int) error {
	if e.contract == nil {
		return fmt.Errorf("runtime not initialized")
	}

	log.Info().Msgf("contract %v", e.contract.Erc20Transactor)

	tx, err := e.contract.Transfer(e.baseDeployer, to, amount)
	if err != nil {
		return fmt.Errorf("failed to transfer tokens: %v", err)
	}

	_, err = bind.WaitMined(context.Background(), e.provider, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transfer transaction: %v", err)
	}

	return nil
}

func (e *ERC20TxBuilder) GetTokenSymbol() string {
	return e.coinSymbol
}

func (e *ERC20TxBuilder) GetValue() *big.Int {
	return e.defaultValue
}

func (e *ERC20TxBuilder) ConstructTransactions(accounts []*tooltypes.SenderAccount, numTx int) ([]*types.Transaction, error) {
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

	log.Info().Msgf("Constructing %s transfer transactions...", e.coinName)
	bar := progressbar.Default(int64(numTx))

	transactions := make([]*types.Transaction, numTx)

	for i := 0; i < numTx; i++ {
		senderIndex := i % len(accounts)
		receiverIndex := (i + 1) % len(accounts)

		sender := accounts[senderIndex]
		receiver := accounts[receiverIndex]

		// log.Info().Msgf("sender %d: %v", i, sender)

		input, err := ContructErc20Transfer(receiver.GetAddress(), e.defaultTransferValue)
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
