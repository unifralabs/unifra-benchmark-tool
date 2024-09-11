package distributor

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"

	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

type TokenRuntimeCosts struct {
	TotalCost  int64
	SubAccount int64
}

type TokenDistributor struct {
	mnemonic             string
	tokenRuntime         tooltypes.Erc20TxBuilder
	totalTx              int
	readyMnemonicIndexes []int
}

func NewTokenDistributor(mnemonic string, readyMnemonicIndexes []int, totalTx int, tokenRuntime tooltypes.Erc20TxBuilder) (*TokenDistributor, error) {
	return &TokenDistributor{
		mnemonic:             mnemonic,
		tokenRuntime:         tokenRuntime,
		totalTx:              totalTx,
		readyMnemonicIndexes: readyMnemonicIndexes,
	}, nil
}

func (td *TokenDistributor) DistributeTokens() ([]int, error) {
	log.Info().Msg("🪙 Token distribution initialized 🪙")

	baseCosts, err := td.calculateRuntimeCosts()
	if err != nil {
		return nil, err
	}
	td.printCostTable(baseCosts)

	shortAddresses, err := td.findAccountsForDistribution(baseCosts.SubAccount)
	if err != nil {
		return nil, err
	}

	initialAccCount := len(shortAddresses)

	if initialAccCount == 0 {
		log.Info().Msg("Accounts are fully funded with tokens for the cycle")
		return td.readyMnemonicIndexes, nil
	}

	fundableAccounts, err := td.getFundableAccounts(baseCosts, shortAddresses)
	if err != nil {
		return nil, err
	}

	if len(fundableAccounts) != initialAccCount {
		log.Info().Msgf("Unable to fund all sub-accounts. Funding %d", len(fundableAccounts))
	}

	err = td.fundAccounts(baseCosts, fundableAccounts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Fund distribution finished!")

	return td.readyMnemonicIndexes, nil
}

func (td *TokenDistributor) calculateRuntimeCosts() (TokenRuntimeCosts, error) {
	transferValue, err := td.tokenRuntime.GetTransferValue()
	if err != nil {
		return TokenRuntimeCosts{}, err
	}

	totalCost := transferValue.Int64() * int64(td.totalTx)
	subAccountCost := int64(math.Ceil(float64(totalCost) / float64(len(td.readyMnemonicIndexes))))

	return TokenRuntimeCosts{
		TotalCost:  totalCost,
		SubAccount: subAccountCost,
	}, nil
}

func (td *TokenDistributor) findAccountsForDistribution(singleRunCost int64) ([]*DistributeAccount, error) {
	log.Info().Msg("Fetching sub-account token balances...")

	shortAddresses := make([]*DistributeAccount, 0)

	bar := progressbar.Default(int64(len(td.readyMnemonicIndexes)))

	for _, index := range td.readyMnemonicIndexes {
		privateKey, err := crypto.HexToECDSA(fmt.Sprintf("%x", crypto.Keccak256([]byte(td.mnemonic), []byte(fmt.Sprintf("m/44'/60'/0'/0/%d", index)))))
		if err != nil {
			return nil, err
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)

		balance, err := td.tokenRuntime.GetTokenBalance(address)
		if err != nil {
			return nil, err
		}

		bar.Add(1)

		if balance.Int64() < singleRunCost {
			shortAddresses = append(shortAddresses, &DistributeAccount{
				MissingFunds:  big.NewInt(singleRunCost - balance.Int64()),
				Address:       address,
				MnemonicIndex: index,
			})
		}
	}

	log.Info().Msg("Fetched initial token balances")

	return shortAddresses, nil
}

func (td *TokenDistributor) printCostTable(costs TokenRuntimeCosts) {
	log.Info().Msg("Cycle Token Cost Table:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", fmt.Sprintf("Cost [%s]", td.tokenRuntime.GetTokenSymbol())})

	table.Append([]string{"Required acc. token balance", fmt.Sprintf("%d", costs.SubAccount)})
	table.Append([]string{"Total token distribution cost", fmt.Sprintf("%d", costs.TotalCost)})

	table.Render()
}

func (td *TokenDistributor) fundAccounts(costs TokenRuntimeCosts, accounts []*DistributeAccount) error {
	log.Info().Msg("Funding accounts with tokens...")

	// Clear the list of ready indexes
	td.readyMnemonicIndexes = []int{}

	bar := progressbar.Default(int64(len(accounts)))

	for _, acc := range accounts {
		err := td.tokenRuntime.FundAccount(acc.Address, acc.MissingFunds)
		if err != nil {
			return err
		}

		bar.Add(1)
		td.readyMnemonicIndexes = append(td.readyMnemonicIndexes, acc.MnemonicIndex)
	}

	return nil
}

func (td *TokenDistributor) getFundableAccounts(costs TokenRuntimeCosts, initialSet []*DistributeAccount) ([]*DistributeAccount, error) {
	distributorBalance, err := td.tokenRuntime.GetSupplierBalance()
	if err != nil {
		return nil, err
	}

	// Sort accounts by missing funds (ascending)
	sort.Slice(initialSet, func(i, j int) bool {
		return initialSet[i].MissingFunds.Cmp(initialSet[j].MissingFunds) < 0
	})

	accountsToFund := []*DistributeAccount{}

	for _, acc := range initialSet {
		if distributorBalance.Cmp(big.NewInt(costs.SubAccount)) > 0 {
			distributorBalance.Sub(distributorBalance, acc.MissingFunds)
			accountsToFund = append(accountsToFund, acc)
		} else {
			break
		}
	}

	if len(accountsToFund) == 0 {
		return nil, fmt.Errorf("not enough funds in distributor")
	}

	return accountsToFund, nil
}

type DistributeAccount struct {
	MissingFunds  *big.Int
	Address       common.Address
	MnemonicIndex int
}
