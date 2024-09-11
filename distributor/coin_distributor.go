package distributor

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	rpc_client "github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

type RuntimeCosts struct {
	AccDistributionCost *big.Int
	SubAccount          *big.Int
}

type Distributor struct {
	ethWallet            *bind.TransactOpts
	mnemonic             string
	provider             *ethclient.Client
	rpcClient            *rpc_client.RpcClient
	runtimeEstimator     tooltypes.TxBuilder
	totalTx              int
	requestedSubAccounts int
	readyMnemonicIndexes []int
}

func NewDistributor(mnemonic string, subAccounts, totalTx int, runtimeEstimator tooltypes.TxBuilder, url string) (*Distributor, error) {

	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc_client.NewRpcClientFromEthClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating eth client: %s", err)
	}

	_, privateKey, err := utils.DerivePrivateKeyFromMnemonic(mnemonic, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	return &Distributor{
		ethWallet:            auth,
		mnemonic:             mnemonic,
		provider:             client,
		rpcClient:            rpcClient,
		runtimeEstimator:     runtimeEstimator,
		totalTx:              totalTx,
		requestedSubAccounts: subAccounts,
		readyMnemonicIndexes: []int{},
	}, nil
}

func (d *Distributor) Distribute() ([]int, error) {
	log.Info().Msg("💸 Fund distribution initialized 💸")

	baseCosts, err := d.calculateRuntimeCosts()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate runtime costs: %v", err)
	}
	d.printCostTable(baseCosts)

	shortAddresses, err := d.findAccountsForDistribution(baseCosts.SubAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts for distribution: %v", err)
	}

	initialAccCount := len(shortAddresses)

	if initialAccCount == 0 {
		log.Info().Msg("Accounts are fully funded for the cycle")
		return d.readyMnemonicIndexes, nil
	}

	fundableAccounts, err := d.getFundableAccounts(baseCosts, shortAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to get fundable accounts: %v", err)
	}

	if len(fundableAccounts) != initialAccCount {
		log.Info().Msgf("Unable to fund all sub-accounts. Funding %d", len(fundableAccounts))
	}

	err = d.fundAccounts(baseCosts, fundableAccounts)
	if err != nil {
		return nil, fmt.Errorf("failed to fund accounts: %v", err)
	}

	log.Info().Msg("Fund distribution finished!")

	return d.readyMnemonicIndexes, nil
}

func (d *Distributor) calculateRuntimeCosts() (*RuntimeCosts, error) {
	inherentValue := d.runtimeEstimator.GetValue()
	baseTxEstimate, err := d.runtimeEstimator.EstimateGasForBaseTx()
	if err != nil {
		return nil, fmt.Errorf("failed to estimate base transaction: %v", err)
	}
	baseGasPrice, err := d.rpcClient.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	baseTxCost := new(big.Int).Mul(baseGasPrice, baseTxEstimate)
	baseTxCost.Add(baseTxCost, inherentValue)

	toAddress, err := utils.DeriveAddressFromMnemonic(d.mnemonic, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %v", err)
	}

	subAccountCost := new(big.Int).Mul(big.NewInt(int64(d.totalTx)), baseTxCost)

	singleDistributionCost, err := d.provider.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  d.ethWallet.From,
		To:    toAddress,
		Value: subAccountCost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate distribution cost: %v", err)
	}

	return &RuntimeCosts{
		AccDistributionCost: big.NewInt(int64(singleDistributionCost)),
		SubAccount:          subAccountCost,
	}, nil
}

func (d *Distributor) findAccountsForDistribution(singleRunCost *big.Int) ([]*DistributeAccount, error) {
	log.Info().Msg("Fetching sub-account balances...")
	bar := progressbar.Default(int64(d.requestedSubAccounts))

	var shortAddresses []*DistributeAccount

	for i := 1; i <= int(d.requestedSubAccounts); i++ {
		address, err := utils.DeriveAddressFromMnemonic(d.mnemonic, int(i))
		if err != nil {
			return nil, err
		}
		balance, err := d.provider.BalanceAt(context.Background(), *address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get balance for address %s: %v", address.Hex(), err)
		}
		bar.Add(1)

		if balance.Cmp(singleRunCost) < 0 {
			shortAddresses = append(shortAddresses, &DistributeAccount{
				MissingFunds:  new(big.Int).Sub(singleRunCost, balance),
				Address:       *address,
				MnemonicIndex: i,
			})
			continue
		}

		d.readyMnemonicIndexes = append(d.readyMnemonicIndexes, i)
	}

	return shortAddresses, nil
}

func (d *Distributor) printCostTable(costs *RuntimeCosts) {
	log.Info().Msg("Cycle Cost Table:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Cost [eth]"})

	table.Append([]string{"Required acc. balance", utils.FormatEther(costs.SubAccount)})
	table.Append([]string{"Single distribution cost", utils.FormatEther(costs.AccDistributionCost)})

	table.Render()
}

func (d *Distributor) getFundableAccounts(costs *RuntimeCosts, initialSet []*DistributeAccount) ([]*DistributeAccount, error) {
	distributorBalance, err := d.provider.BalanceAt(context.Background(), d.ethWallet.From, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get distributor balance: %v", err)
	}

	var accountsToFund []*DistributeAccount

	sort.Slice(initialSet, func(i, j int) bool {
		return initialSet[i].MissingFunds.Cmp(initialSet[j].MissingFunds) < 0
	})

	for _, acc := range initialSet {
		if distributorBalance.Cmp(costs.AccDistributionCost) > 0 {
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

func (d *Distributor) fundAccounts(costs *RuntimeCosts, accounts []*DistributeAccount) error {
	log.Info().Msg("Funding accounts...")
	bar := progressbar.Default(int64(len(accounts)))

	for _, acc := range accounts {
		log.Info().Msgf("Funding account %s with coins %s", acc.Address.Hex(), utils.FormatEther(acc.MissingFunds))
		tx, err := d.rpcClient.BuildTransferTx(acc.Address, acc.MissingFunds, costs.AccDistributionCost.Uint64(), d.ethWallet)
		if err != nil {
			return err
		}

		_, err = d.rpcClient.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("failed to send transaction to %s: %v", acc.Address.Hex(), err)
		}

		_, err = bind.WaitMined(context.Background(), d.provider, tx)
		if err != nil {
			return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
		}

		bar.Add(1)
		d.readyMnemonicIndexes = append(d.readyMnemonicIndexes, acc.MnemonicIndex)
	}

	return nil
}
