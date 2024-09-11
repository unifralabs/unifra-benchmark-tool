package types

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
)

type SenderAccount struct {
	MnemonicIndex int
	Nonce         uint64
	Wallet        *accounts.Account
	Address       common.Address
	PrivateKey    *ecdsa.PrivateKey
}

func NewSenderAccount(mnemonicIndex int, nonce uint64, wallet *accounts.Account, privateKey *ecdsa.PrivateKey) (*SenderAccount, error) {

	return &SenderAccount{
		MnemonicIndex: mnemonicIndex,
		Nonce:         nonce,
		Wallet:        wallet,
		PrivateKey:    privateKey,
	}, nil
}

func (sa *SenderAccount) IncrNonce() {
	sa.Nonce++
}

func (sa *SenderAccount) GetNonce() uint64 {
	return sa.Nonce
}

func (sa *SenderAccount) GetAddress() common.Address {
	return sa.Wallet.Address
}
