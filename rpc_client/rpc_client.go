package rpc_client

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RpcClient struct {
	client  *ethclient.Client
	chainID *big.Int
}

func NewRpcClient(rpcUrl string) (*RpcClient, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	return &RpcClient{client: client, chainID: chainID}, nil
}

func NewRpcClientFromEthClient(client *ethclient.Client) (*RpcClient, error) {

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	return &RpcClient{client: client, chainID: chainID}, nil
}

func (e *RpcClient) GetBalance(address string) (uint64, error) {
	balance, err := e.client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return 0, err
	}

	return balance.Uint64(), nil
}

func (e *RpcClient) GetNonce(address string) (uint64, error) {
	nonce, err := e.client.NonceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (e *RpcClient) GetCode(address string) (string, error) {
	code, err := e.client.CodeAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(code), nil
}

func (e *RpcClient) GetStorageAt(address string, index string) (string, error) {
	indexBigInt := new(big.Int)
	indexBigInt.SetString(index, 0)
	storage, err := e.client.StorageAt(context.Background(), common.HexToAddress(address), common.BigToHash(indexBigInt), nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(storage), nil
}

func (e *RpcClient) GetGasPrice() (*big.Int, error) {
	gasPrice, err := e.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	return gasPrice, nil
}

func (ec *RpcClient) BuildTransferTx(to common.Address, value *big.Int, gasLimit uint64, opts *bind.TransactOpts) (tx *types.Transaction, err error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = ec.client.PendingNonceAt(ctx, opts.From)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}

	if opts.GasPrice == nil {
		price, err := ec.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
		opts.GasPrice = price
	}

	baseTx := types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		GasPrice: opts.GasPrice,
		Gas:      gasLimit,
		Value:    value,
		Data:     []byte{},
	}
	rawTx := types.NewTx(&baseTx)

	signedTx, err := opts.Signer(opts.From, rawTx)
	if err != nil {
		err = errors.WithMessage(err, "signed raw tx")
		return
	}
	return signedTx, nil
}

// Send a transaction to the ETH network
func (e *RpcClient) GenerateRawTransaction(fromAddress string, toAddress string, amount uint64, privateKeyHex string) (string, error) {
	nonce, err := e.GetNonce(fromAddress)
	if err != nil {
		return "", err
	}

	amountBigInt := new(big.Int).SetUint64(amount)
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddress), amountBigInt, 21000, big.NewInt(1000000000), nil)

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}

	tx, err = types.SignTx(tx, types.NewEIP155Signer(e.chainID), privateKey)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

// EstimateGasForTransfer estimates the gas needed for a native token transfer transaction
func (e *RpcClient) EstimateGasForTransfer(fromAddress string, toAddress string, amount uint64) (uint64, error) {
	toAddr := common.HexToAddress(toAddress)
	gas, err := e.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  common.HexToAddress(fromAddress),
		To:    &toAddr,
		Value: new(big.Int).SetUint64(amount),
	})
	if err != nil {
		return 0, err
	}

	return gas, nil
}

// EstimateGasForContractCall estimates the gas needed for a contract-call transaction
func (e *RpcClient) EstimateGasForContractCall(fromAddress common.Address, toAddress common.Address, data []byte) (uint64, error) {
	gas, err := e.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &toAddress,
		Data: data,
	})
	if err != nil {
		return 0, err
	}

	return gas, nil
}

// SendRawTransaction sends a raw transaction to the ETH network
func (e *RpcClient) SendRawTransaction(rawTx string) (string, error) {
	tx, err := hexutil.Decode(rawTx)
	if err != nil {
		return "", err
	}

	decodedTx := new(types.Transaction)
	err = decodedTx.UnmarshalBinary(tx)
	if err != nil {
		return "", err
	}

	err = e.client.SendTransaction(context.Background(), decodedTx)
	if err != nil {
		return "", err
	}

	return decodedTx.Hash().Hex(), nil
}

func (e *RpcClient) SendTransaction(tx *types.Transaction) (string, error) {
	err := e.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

// web3_client_version
func (e *RpcClient) GetClientVersion() (string, error) {
	var result string
	err := e.client.Client().Call(&result, "web3_clientVersion")
	if err != nil {
		return "", err
	}
	return result, nil
}
