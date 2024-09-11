package rpc_builder

import (
	"math/rand"
	"time"

	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

// GenerateTransactionHashes generates a sequence of transaction hashes
func GenerateTransactionHashes(n int, network *string, randomSeed *tooltypes.RandomSeed) []string {
	return LoadSamples(LoadSamplesParams{
		Network:    network,
		Datatype:   "transactions",
		N:          n,
		RandomSeed: randomSeed,
	})
}

// LoadSamplesParams represents the parameters for LoadSamples
type LoadSamplesParams struct {
	Network    *string
	Datatype   string
	N          int
	RandomSeed *tooltypes.RandomSeed
}

// LoadSamples loads sample data
func LoadSamples(params LoadSamplesParams) []string {
	// Initialize random number generator
	var r *rand.Rand
	if params.RandomSeed != nil {
		r = rand.New(rand.NewSource(int64(*params.RandomSeed)))
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// TODO: Implement actual sample loading logic
	// This is a placeholder implementation
	samples := make([]string, params.N)
	for i := 0; i < params.N; i++ {
		samples[i] = generateRandomHash(r)
	}

	return samples
}

// generateRandomHash generates a random hash-like string
func generateRandomHash(r *rand.Rand) string {
	const charset = "abcdef0123456789"
	hash := make([]byte, 64)
	for i := range hash {
		hash[i] = charset[r.Intn(len(charset))]
	}
	return "0x" + string(hash)
}

// GenerateContractAddresses generates a sequence of contract addresses
func GenerateContractAddresses(n int, network *string, randomSeed *tooltypes.RandomSeed) ([]string, error) {
	return LoadAddressSamples(LoadSamplesParams{
		Network:    network,
		Datatype:   "contracts",
		N:          n,
		RandomSeed: randomSeed,
	}), nil
}

// GenerateEOAs generates a sequence of externally owned account addresses
func GenerateEOAs(n int, network *string, randomSeed *tooltypes.RandomSeed) ([]string, error) {
	return LoadAddressSamples(LoadSamplesParams{
		Network:    network,
		Datatype:   "eoas",
		N:          n,
		RandomSeed: randomSeed,
	}), nil
}

// LoadSamples loads sample data
func LoadAddressSamples(params LoadSamplesParams) []string {
	// TODO: Implement actual sample loading logic
	// This is a placeholder implementation
	samples := make([]string, params.N)
	for i := 0; i < params.N; i++ {
		samples[i] = generateRandomAddress()
	}
	return samples
}

// generateRandomAddress generates a random Ethereum-like address
func generateRandomAddress() string {
	// Placeholder implementation
	return "0x" + generateRandomHexString(40)
}

// generateRandomHexString generates a random hex string of given length
func generateRandomHexString(length int) string {
	const charset = "0123456789abcdef"
	result := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}
	return string(result)
}
