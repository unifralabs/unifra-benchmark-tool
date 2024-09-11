package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxBuilder interface {
	// Estimates the base runtime transaction gas limit
	EstimateGasForBaseTx() (*big.Int, error)

	// Returns the value of each cycle transaction, if any
	GetValue() *big.Int

	// Constructs the specific runtime transactions
	ConstructTransactions(accounts []*SenderAccount, numTxs int) ([]*types.Transaction, error)

	// Initializes the runtime
	Initialize() error
}

type Erc20TxBuilder interface {
	TxBuilder
	// Returns the transfer token value
	GetTransferValue() (*big.Int, error)

	// Returns the token balance for the specified address
	GetTokenBalance(address common.Address) (*big.Int, error)

	// Returns the suppliers token balance
	GetSupplierBalance() (*big.Int, error)

	// Returns the token name
	GetTokenSymbol() string

	// Funds the specified account
	FundAccount(address common.Address, amount *big.Int) error
}

type Erc721TxBuilder interface {
	TxBuilder
}
