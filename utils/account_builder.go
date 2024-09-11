package utils

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

func GetSenderAccounts(ethclient *ethclient.Client, mnemonic string, accountIndexes []int, numTxs int) ([]*tooltypes.SenderAccount, error) {
	log.Info().Msg("Gathering initial account nonces...")

	walletsToInit := len(accountIndexes)
	if numTxs < walletsToInit {
		walletsToInit = numTxs
	}

	bar := progressbar.Default(int64(walletsToInit))

	accounts := make([]*tooltypes.SenderAccount, 0, walletsToInit)
	for i := 0; i < walletsToInit; i++ {
		accIndex := accountIndexes[i]

		account, privateKey, err := DerivePrivateKeyFromMnemonic(mnemonic, accIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet: %v", err)
		}

		nonce, err := ethclient.PendingNonceAt(context.Background(), account.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %v", err)
		}

		senderAccount, err := tooltypes.NewSenderAccount(accIndex, nonce, account, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create sender account: %v", err)
		}

		accounts = append(accounts, senderAccount)

		bar.Add(1)
	}

	log.Info().Msg("✅ Gathered initial nonce data")

	return accounts, nil
}
