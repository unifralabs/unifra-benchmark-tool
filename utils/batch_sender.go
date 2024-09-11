package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

func BatchSendRawTransactions(batches [][]*types.Transaction, url string) ([]string, error) {

	log.Info().Msg("Sending transactions in batches...")

	bar := progressbar.Default(int64(len(batches)))

	txHashes := []string{}
	batchErrors := []string{}

	client := &http.Client{}

	for _, batch := range batches {
		var singleRequests []*tooltypes.JsonrpcMessage
		for _, signedTx := range batch {
			signedTxBytes, err := signedTx.MarshalBinary()
			if err != nil {
				batchErrors = append(batchErrors, fmt.Sprintf("failed to marshal tx: %v", err))
				continue
			}

			singleRequests = append(singleRequests, NewJsonrpcMessage("eth_sendRawTransaction", []interface{}{hexutil.Encode(signedTxBytes)}))
		}

		requestBody, err := json.Marshal(singleRequests)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}

		// log.Info().Msgf("request: %s", string(requestBody))

		req, err := http.NewRequestWithContext(context.Background(), "POST", url, strings.NewReader(string(requestBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		var responses []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
		// log.Info().Msgf("responses: %s", responses)

		for _, response := range responses {
			if err, ok := response["error"]; ok {
				log.Info().Msgf("error: %s", err)

				batchErrors = append(batchErrors, fmt.Sprintf("%v", err))
			} else if result, ok := response["result"].(string); ok {
				txHashes = append(txHashes, result)
			}
		}

		bar.Add(1)
	}

	if len(batchErrors) > 0 {
		log.Info().Msg("Errors encountered during batch sending:")
		for _, err := range batchErrors {
			log.Info().Msg(err)
		}
	}

	log.Info().Msgf("Batches sent: %d", len(batches))

	return txHashes, nil
}
