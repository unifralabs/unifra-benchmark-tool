package utils

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

func SignTransactions(ethclient *ethclient.Client, accounts []*tooltypes.SenderAccount, transactions []*types.Transaction) ([]*types.Transaction, error) {
	failedTxnSignErrors := make([]error, 0)

	log.Info().Msg("Signing transactions...")

	bar := progressbar.Default(int64(len(transactions)))
	signedTxs := make([]*types.Transaction, 0, len(transactions))

	chainID, err := ethclient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	for i, tx := range transactions {
		sender := accounts[i%len(accounts)]

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), sender.PrivateKey)
		if err != nil {
			failedTxnSignErrors = append(failedTxnSignErrors, err)
			continue
		}

		signedTxs = append(signedTxs, signedTx)

		bar.Add(1)
	}

	log.Info().Msgf("✅ Successfully signed %d transactions", len(signedTxs))

	if len(failedTxnSignErrors) > 0 {
		log.Warn().Msg("Errors encountered during transaction signing:")
		for _, err := range failedTxnSignErrors {
			log.Error().Msg(err.Error())
		}
	}

	return signedTxs, nil
}
