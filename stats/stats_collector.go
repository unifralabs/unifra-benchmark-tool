package stats

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

type TxStats struct {
	TxHash string
	Block  uint64
}

type BlockInfo struct {
	BlockNum       uint64
	CreatedAt      uint64
	NumTxs         int
	GasUsed        uint64
	GasLimit       uint64
	GasUtilization float64
}

type CollectorData struct {
	TPS       float64
	BlockInfo map[uint64]*BlockInfo
}

func GatherTransactionReceipts(ethclient *ethclient.Client, txs []*types.Transaction, batchSize int) ([]*TxStats, error) {
	log.Info().Msg("Gathering transaction receipts...")
	bar := progressbar.Default(int64(len(txs)))

	var wg sync.WaitGroup
	txStatsChan := make(chan *TxStats, len(txs))
	errorsChan := make(chan error, len(txs))

	for i := 0; i < len(txs); i += int(batchSize) {
		end := i + int(batchSize)
		if end > len(txs) {
			end = len(txs)
		}
		batch := txs[i:end]

		wg.Add(1)
		go func(batch []*types.Transaction) {
			defer wg.Done()
			for _, tx := range batch {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
				receipt, err := bind.WaitMined(ctx, ethclient, tx)
				cancel()

				if err != nil {
					errorsChan <- err
					continue
				}
				txStatsChan <- &TxStats{TxHash: tx.Hash().Hex(), Block: receipt.BlockNumber.Uint64()}
				bar.Add(1)
			}
		}(batch)
	}

	go func() {
		wg.Wait()
		close(txStatsChan)
		close(errorsChan)
	}()

	var txStats []*TxStats
	for txStat := range txStatsChan {
		txStats = append(txStats, txStat)
	}

	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		log.Info().Msg("Errors encountered during batch sending:")
		for _, err := range errors {
			log.Info().Msgf("Error: %v", err)
		}
	}

	log.Info().Msg("Gathered transaction receipts")
	return txStats, nil
}

func FetchBlockInfo(ethclient *ethclient.Client, stats []*TxStats) (map[uint64]*BlockInfo, error) {
	log.Info().Msg("Gathering block info...")
	blockSet := make(map[uint64]bool)
	for _, s := range stats {
		blockSet[s.Block] = true
	}

	bar := progressbar.Default(int64(len(blockSet)))

	var wg sync.WaitGroup
	blockInfoChan := make(chan *BlockInfo, len(blockSet))
	errorsChan := make(chan error, len(blockSet))

	for block := range blockSet {
		wg.Add(1)
		go func(blockNum uint64) {
			defer wg.Done()
			blockInfo, err := ethclient.BlockByNumber(context.Background(), big.NewInt(int64(blockNum)))
			if err != nil {
				errorsChan <- err
				return
			}
			gasUtilization := float64(blockInfo.GasUsed()) / float64(blockInfo.GasLimit()) * 100
			blockInfoChan <- &BlockInfo{
				BlockNum:       blockNum,
				CreatedAt:      blockInfo.Time(),
				NumTxs:         len(blockInfo.Transactions()),
				GasUsed:        blockInfo.GasUsed(),
				GasLimit:       blockInfo.GasLimit(),
				GasUtilization: gasUtilization,
			}
			bar.Add(1)
		}(block)
	}

	go func() {
		wg.Wait()
		close(blockInfoChan)
		close(errorsChan)
	}()

	blocksMap := make(map[uint64]*BlockInfo)
	for blockInfo := range blockInfoChan {
		blocksMap[blockInfo.BlockNum] = blockInfo
	}

	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		log.Info().Msg("Errors encountered during block info fetch:")
		for _, err := range errors {
			log.Info().Msgf("Error: %v", err)
		}
	}

	log.Info().Msg("Gathered block info")
	return blocksMap, nil
}

func CalcTPS(stats []*TxStats, blockInfoMap map[uint64]*BlockInfo) float64 {
	log.Info().Msg("🧮 Calculating TPS data 🧮")
	totalTxs := 0
	totalTime := uint64(0)

	var blocks []uint64
	for block := range blockInfoMap {
		blocks = append(blocks, block)
	}
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })

	for i := 1; i < len(blocks); i++ {
		currentBlock := blockInfoMap[blocks[i]]
		prevBlock := blockInfoMap[blocks[i-1]]
		totalTxs += currentBlock.NumTxs
		totalTime += currentBlock.CreatedAt - prevBlock.CreatedAt
	}

	if totalTime == 0 {
		return 0
	}
	return math.Ceil(float64(totalTxs) / float64(totalTime))
}

func PrintBlockData(blockInfoMap map[uint64]*BlockInfo) {
	log.Info().Msg("Block utilization data:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Block #", "Gas Used [wei]", "Gas Limit [wei]", "Transactions", "Utilization"})

	var blocks []uint64
	for block := range blockInfoMap {
		blocks = append(blocks, block)
	}
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })

	for _, block := range blocks {
		info := blockInfoMap[block]
		table.Append([]string{
			fmt.Sprintf("%d", info.BlockNum),
			fmt.Sprintf("%d", info.GasUsed),
			fmt.Sprintf("%d", info.GasLimit),
			fmt.Sprintf("%d", info.NumTxs),
			fmt.Sprintf("%.2f%%", info.GasUtilization),
		})
	}

	table.Render()
}

func PrintFinalData(tps float64, blockInfoMap map[uint64]*BlockInfo) {
	totalUtilization := 0.0
	for _, info := range blockInfoMap {
		totalUtilization += info.GasUtilization
	}
	avgUtilization := totalUtilization / float64(len(blockInfoMap))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TPS", "Blocks", "Avg. Utilization"})
	table.Append([]string{
		fmt.Sprintf("%d", int(tps)),
		fmt.Sprintf("%d", len(blockInfoMap)),
		fmt.Sprintf("%.2f%%", avgUtilization),
	})
	table.Render()
}

func GenerateStats(ethclient *ethclient.Client, txHashes []*types.Transaction, batchSize int) (*CollectorData, error) {
	if len(txHashes) == 0 {
		log.Info().Msg("No stat data to display")
		return &CollectorData{TPS: 0, BlockInfo: make(map[uint64]*BlockInfo)}, nil
	}

	log.Info().Msg("⏱ Statistics calculation initialized ⏱")

	txStats, err := GatherTransactionReceipts(ethclient, txHashes, batchSize)
	if err != nil {
		return nil, err
	}

	blockInfoMap, err := FetchBlockInfo(ethclient, txStats)
	if err != nil {
		return nil, err
	}

	PrintBlockData(blockInfoMap)

	avgTPS := CalcTPS(txStats, blockInfoMap)
	PrintFinalData(avgTPS, blockInfoMap)

	return &CollectorData{TPS: avgTPS, BlockInfo: blockInfoMap}, nil
}
