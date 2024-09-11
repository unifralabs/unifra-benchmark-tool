package config

import "github.com/spf13/viper"

type EnvConfig struct {
	TestName                 string `mapstructure:"TEST_NAME"`
	NodeName                 string `mapstructure:"NODE_NAME"`
	NumTestAccounts          int    `mapstructure:"NUM_TEST_ACCOUNTS"`
	AdminAccountMnemonic     string `mapstructure:"ADMIN_ACCOUNT_MNEMONIC"`
	RpcUrl                   string `mapstructure:"RPC_URL"`
	OutputDir                string `mapstructure:"OUTPUT_DIR"`
	SendTransactionBatchSize int    `mapstructure:"SEND_TRANSACTION_BATCH_SIZE"`
}

// Load config file via viper
func LoadEnvConfig() (*EnvConfig, error) {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	cfg := EnvConfig{}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
