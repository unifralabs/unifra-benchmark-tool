package types

import (
	"errors"
)

// EstimateCallCount calculates the total number of calls for a load test
func EstimateCallCount(rates []int, durations []int, nRepeats *int) (int, error) {
	var nCalls int

	if durations != nil {
		if len(rates) != len(durations) {
			return 0, errors.New("different number of rates vs durations")
		}
		for i, rate := range rates {
			nCalls += rate * durations[i]
		}
	} else {
		return 0, errors.New("should specify duration or durations")
	}

	if nRepeats != nil {
		nCalls *= *nRepeats
	}

	return nCalls, nil
}

// CreateLoadTest creates a load test configuration
func CreateLoadTest(calls []*JsonrpcMessage, rates []int, durations []int, vegetaArgs interface{}, repeatCalls bool) ([]VegetaAttack, error) {
	// Validate inputs
	if len(rates) == 0 {
		return nil, errors.New("must specify at least one rate")
	}

	// Pluralize singular durations
	useDurations := durations
	if len(useDurations) != len(rates) {
		return nil, errors.New("number of durations must match number of rates")
	}

	// Pluralize singular vegeta args
	useVegetaArgs := make([]*string, len(rates))
	switch v := vegetaArgs.(type) {
	case nil:
		for i := range useVegetaArgs {
			useVegetaArgs[i] = nil
		}
	case string:
		for i := range useVegetaArgs {
			useVegetaArgs[i] = &v
		}
	default:
		return nil, errors.New("invalid vegeta args input")
	}

	// Partition calls into individual attacks
	var attacksCalls [][]*JsonrpcMessage
	if !repeatCalls {
		callsIter := calls
		for i, rate := range rates {
			nAttackCalls := rate * useDurations[i]
			attackCalls := make([]*JsonrpcMessage, nAttackCalls)
			for j := 0; j < nAttackCalls; j++ {
				if len(callsIter) == 0 {
					return nil, errors.New("not enough calls provided")
				}
				attackCalls[j] = callsIter[0]
				callsIter = callsIter[1:]
			}
			attacksCalls = append(attacksCalls, attackCalls)
		}
	} else {
		for range rates {
			attacksCalls = append(attacksCalls, calls)
		}
	}

	// Create load tests
	loadTest := make([]VegetaAttack, len(rates))
	for i, rate := range rates {
		loadTest[i] = VegetaAttack{
			Rate:       rate,
			Duration:   useDurations[i],
			Calls:      attacksCalls[i],
			VegetaArgs: useVegetaArgs[i],
		}
	}

	return loadTest, nil
}
