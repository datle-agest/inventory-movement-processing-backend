package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	movements := generateFakeData(100)

	var (
		accepted  int32
		rejected  int32
		duplicate int32
		seen      = make(map[string]bool)
		mu        sync.Mutex
		wg        sync.WaitGroup
		jobs      = make(chan mockMovement, len(movements))
	)

	// 10 workers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				mu.Lock()
				if seen[m.ID] {
					mu.Unlock()
					atomic.AddInt32(&duplicate, 1)
					continue
				}
				seen[m.ID] = true
				mu.Unlock()

				if err := m.validate(); err != nil {
					atomic.AddInt32(&rejected, 1)
					continue
				}
				atomic.AddInt32(&accepted, 1)
			}
		}()
	}

	for _, m := range movements {
		jobs <- m
	}
	close(jobs)
	wg.Wait()

	out, _ := json.Marshal(result{
		AcceptedCount:  accepted,
		RejectedCount:  rejected,
		DuplicateCount: duplicate,
	})
	fmt.Println(string(out))
}

func generateFakeData(n int) []mockMovement {
	types := []movementType{movementTypeIn, movementTypeOut, movementTypeAdjust}
	movements := make([]mockMovement, n)

	for i := 0; i < n; i++ {
		movements[i] = mockMovement{
			ID:       fmt.Sprintf("MOV-%03d", i),
			Name:     fmt.Sprintf("movement %d", i),
			ItemID:   int32(i%10 + 1),
			Type:     types[i%3],
			Quantity: int32(i%50 + 1),
		}

		switch {
		case i%15 == 0: // rejected
			movements[i].Quantity = -1
		case i%10 == 0: // duplicate
			movements[i].ID = "MOV-000"
		}
	}
	return movements
}
