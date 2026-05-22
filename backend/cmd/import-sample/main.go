package main

func main() {

}

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	movementEntity "inventory-movement-processing/internal/movement/entity"
//	"inventory-movement-processing/internal/movement/service"
//	"inventory-movement-processing/pkg/components/workerc"
//	"log"
//	"os"
//	"sync"
//	"sync/atomic"
//	"time"
//)
//
//func main() {
//	movements, err := readFromJSON("cmd/import-sample/testdata/movements.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println("========== SEQUENTIAL (no pool) ==========")
//	uc0 := service.NewMovementService(nil)
//	start := time.Now()
//	var seq importResult
//	for _, m := range movements {
//		m := m
//		switch uc0.ProcessOne(context.Background(), &m) {
//		case service.StatusAccepted:
//			seq.AcceptedCount++
//		case service.StatusRejected:
//			seq.RejectedCount++
//		case service.StatusDuplicate:
//			seq.DuplicateCount++
//		}
//	}
//	fmt.Printf("Took: %v\n", time.Since(start))
//	out, _ := json.MarshalIndent(seq, "", "  ")
//	fmt.Println(string(out))
//
//	fmt.Println("========== 5 WORKERS (pool) ==========")
//	uc5 := service.NewMovementService(nil)
//	r5, _ := json.MarshalIndent(runWorkerPool(movements, uc5, 5), "", "  ")
//	fmt.Println(string(r5))
//
//	fmt.Println("========== 10 WORKERS (pool) ==========")
//	uc10 := service.NewMovementService(nil)
//	r10, _ := json.MarshalIndent(runWorkerPool(movements, uc10, 10), "", "  ")
//	fmt.Println(string(r10))
//}
//
//type importResult struct {
//	AcceptedCount  int32 `json:"accepted_count"`
//	RejectedCount  int32 `json:"rejected_count"`
//	DuplicateCount int32 `json:"duplicate_count"`
//}
//
//func runWorkerPool(movements []movementEntity.Movement, uc service.MovementService, numWorkers int) importResult {
//	pool := workerc.NewPool("import-pool", numWorkers, len(movements))
//	pool.InitFlags()
//	pool.Activate(nil)
//
//	var accepted, rejected, duplicate atomic.Int32
//	var batchWg sync.WaitGroup
//
//	workerJobCounts := make([]atomic.Int32, numWorkers)
//
//	start := time.Now()
//
//	for i, m := range movements {
//		batchWg.Add(1)
//		m := m
//		idx := i % numWorkers
//		pool.Submit(func() {
//			defer batchWg.Done()
//			workerJobCounts[idx].Add(1)
//			switch uc.ProcessOne(context.Background(), &m) {
//			case service.StatusAccepted:
//				accepted.Add(1)
//			case service.StatusRejected:
//				rejected.Add(1)
//			case service.StatusDuplicate:
//				duplicate.Add(1)
//			}
//		})
//	}
//
//	batchWg.Wait()
//	pool.Stop()
//
//	fmt.Printf("(%d workers, took %v)\n", numWorkers, time.Since(start))
//
//	return importResult{
//		AcceptedCount:  accepted.Load(),
//		RejectedCount:  rejected.Load(),
//		DuplicateCount: duplicate.Load(),
//	}
//}
//
//func readFromJSON(path string) ([]movementEntity.Movement, error) {
//	data, err := os.ReadFile(path)
//	if err != nil {
//		return nil, err
//	}
//
//	var movements []movementEntity.Movement
//	if err := json.Unmarshal(data, &movements); err != nil {
//		return nil, err
//	}
//
//	return movements, nil
//}
