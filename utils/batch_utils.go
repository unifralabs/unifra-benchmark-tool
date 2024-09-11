package utils

func GenerateBatches[ItemType any](items []ItemType, batchSize int) [][]ItemType {
	batches := [][]ItemType{}

	numBatches := (len(items) + batchSize - 1) / batchSize

	for i := 0; i < numBatches; i++ {
		batches = append(batches, []ItemType{})
	}

	currentBatch := 0
	for _, item := range items {
		batches[currentBatch] = append(batches[currentBatch], item)

		if len(batches[currentBatch])%batchSize == 0 {
			currentBatch++
		}
	}

	return batches
}
