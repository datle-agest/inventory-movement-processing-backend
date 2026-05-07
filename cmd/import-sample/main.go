package main

import (
	"context"
	"encoding/json"
	"fmt"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	movementUsecase "inventory-movement-processing/internal/movement/usecase"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

func main() {
	movements, err := readFromJSON("cmd/import-sample/testdata/movements.json")
	if err != nil {
		log.Fatal(err)
	}

	uc := movementUsecase.NewMovementUsecase()
	result := runWorkerPool(movements, uc, 5)

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}

type importResult struct {
	AcceptedCount  int32 `json:"accepted_count"`
	RejectedCount  int32 `json:"rejected_count"`
	DuplicateCount int32 `json:"duplicate_count"`
}

func runWorkerPool(movements []movementEntity.Movement, uc movementUsecase.MovementUsecase, numWorkers int) importResult {
	jobs := make(chan movementEntity.Movement, len(movements))

	var accepted, rejected, duplicate atomic.Int32
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				m := m
				switch uc.ProcessOne(context.Background(), &m) {
				case movementUsecase.StatusAccepted:
					accepted.Add(1)
				case movementUsecase.StatusRejected:
					rejected.Add(1)
				case movementUsecase.StatusDuplicate:
					duplicate.Add(1)
				}
			}
		}()
	}

	for _, m := range movements {
		jobs <- m
	}
	close(jobs)
	wg.Wait()

	return importResult{
		AcceptedCount:  accepted.Load(),
		RejectedCount:  rejected.Load(),
		DuplicateCount: duplicate.Load(),
	}
}

func readFromJSON(path string) ([]movementEntity.Movement, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var movements []movementEntity.Movement
	if err := json.Unmarshal(data, &movements); err != nil {
		return nil, err
	}

	return movements, nil
}
