package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/benchmarker"
	"github.com/unifralabs/unifra-benchmark-tool/config"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func main() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	cfg, err := config.LoadEnvConfig()
	if err != nil {
		log.Info().Msgf("Error loading config: %s", err)
		return
	}

	_, ecdsaPrivateKey, err := utils.DerivePrivateKeyFromMnemonic(cfg.AdminAccountMnemonic, 0)
	if err != nil {
		log.Info().Msgf("failed to derive private key: %v", err)
		return
	}
	publicKey := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	log.Info().Msgf("Admin Account Address %s", publicKey.Hex())
	// privateKey := hex.EncodeToString(crypto.FromECDSA(ecdsaPrivateKey))

	log.Info().Msgf("Config loaded: %v", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	// Notify the sigCh channel when the program receives the interrupt (Ctrl+C) or termination signal.
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		<-sigCh
		cancel()
	}()

	benchmarker, err := benchmarker.NewBenchmarker(cfg)
	if err != nil {
		log.Info().Msgf("Error creating Benchmarker object: %s", err)
		return
	}
	err = benchmarker.Initialize()
	if err != nil {
		log.Info().Msgf("Error initializing Benchmarker: %s", err)
		return
	}
	benchmarker.RunBenchmarks(ctx)
}
