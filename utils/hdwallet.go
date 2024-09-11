package utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// Helper functions
func DeriveAddress(wallet *hdwallet.Wallet, index int) (common.Address, error) {
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	account, err := wallet.Derive(path, true)
	if err != nil {
		return common.Address{}, err
	}
	return account.Address, nil
}

func DerivePrivateKey(wallet *hdwallet.Wallet, index int) (*accounts.Account, *ecdsa.PrivateKey, error) {
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	account, err := wallet.Derive(path, true)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, nil, err
	}
	return &account, privateKey, nil
}

func DeriveAddressFromMnemonic(mnemonic string, index int) (*common.Address, error) {
	// Create a wallet from the mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	address, err := DeriveAddress(wallet, index)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}
	return &address, nil
}

func DerivePrivateKeyFromMnemonic(mnemonic string, index int) (*accounts.Account, *ecdsa.PrivateKey, error) {
	// Create a wallet from the mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return DerivePrivateKey(wallet, index)
}
