package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"time"
)

// ProcessItemGroup processes all movements for a single item ID inside a single transaction.
// It performs row-level locking, sequential simulation, and bulk inserts.
func (s *service) ProcessItemGroup(ctx context.Context, itemID int32, rows []entity.CsvMovementRow) []entity.ProcessResult {
	var results []entity.ProcessResult

	// Set context timeout to prevent hanging transactions and auto-release row locks
	txCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.txManager.WithTx(txCtx, func(txCtx context.Context) error {
		// Set PostgreSQL lock timeout to fail fast during heavy contention
		if err := s.movementRepo.SetLockTimeout(txCtx, "3s"); err != nil {
			s.logger.Errorf("[Service][ProcessItemGroup] failed to set lock timeout: %v", err)
			return err
		}

		// 1. Lock Acquisition
		item, err := s.lockItemForProcessing(txCtx, itemID)
		if err != nil {
			return err
		}

		// 2. Duplicate Pre-check
		externalIDs := collectExternalIDs(rows)
		existingMap, err := s.checkExistingDuplicates(txCtx, externalIDs)
		if err != nil {
			return err
		}

		// 3. Sequential In-Memory Simulation
		finalStock, successfulMovements, simResults := s.simulateMovements(item.CurrentStock, rows, existingMap)
		results = simResults

		// 4. Empty Success Optimization
		if len(successfulMovements) == 0 {
			return nil
		}

		// 5. Final Stock Update
		if err := s.updateFinalStock(txCtx, itemID, finalStock); err != nil {
			return err
		}

		// 6. Bulk Insert
		if err := s.bulkInsertMovements(txCtx, successfulMovements); err != nil {
			return err
		}

		return nil
	})

	// 7. System/DB Failure Handling
	if err != nil {
		return s.createSystemFailureResults(rows, err)
	}

	return results
}

func (s *service) lockItemForProcessing(ctx context.Context, itemID int32) (*itemEntity.Item, error) {
	item, err := s.itemService.GetItemForUpdate(ctx, itemID)
	if err != nil {
		s.logger.Errorf("[Service][ProcessItemGroup] failed to lock item %d: %v", itemID, err)
		return nil, err
	}
	if item == nil {
		s.logger.Warnf("[Service][ProcessItemGroup] item with id %d not found", itemID)
		return nil, common.NewNotFoundError(common.CodeItemNotFound, "item not found")
	}
	return item, nil
}

// collectExternalIDs is a pure function — no service dependency needed
func collectExternalIDs(rows []entity.CsvMovementRow) []string {
	externalIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		externalIDs = append(externalIDs, r.ExternalID)
	}
	return externalIDs
}

func (s *service) checkExistingDuplicates(ctx context.Context, externalIDs []string) (map[string]bool, error) {
	existingDBExternalIDs, err := s.movementRepo.GetExistingExternalIDs(ctx, externalIDs)
	if err != nil {
		s.logger.Errorf("[Service][ProcessItemGroup] failed to check duplicate external_ids: %v", err)
		return nil, err
	}

	existingMap := make(map[string]bool, len(existingDBExternalIDs))
	for _, id := range existingDBExternalIDs {
		existingMap[id] = true
	}
	return existingMap, nil
}

func (s *service) updateFinalStock(ctx context.Context, itemID int32, finalStock int32) error {
	if err := s.itemService.UpdateStock(ctx, itemID, finalStock); err != nil {
		s.logger.Errorf("[Service][ProcessItemGroup] failed to update stock for item %d: %v", itemID, err)
		return err
	}
	return nil
}

func (s *service) bulkInsertMovements(ctx context.Context, movements []*entity.Movement) error {
	if err := s.movementRepo.CreateBatch(ctx, movements); err != nil {
		s.logger.Errorf("[Service][ProcessItemGroup] failed to bulk insert movements: %v", err)
		return err
	}
	return nil
}

func (s *service) createSystemFailureResults(rows []entity.CsvMovementRow, err error) []entity.ProcessResult {
	var failedResults []entity.ProcessResult

	// Default to internal server error if it is not an AppError
	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		appErr = common.ErrInternal("cannot process movement, please try again")
	}

	for _, r := range rows {
		failedResults = append(failedResults, entity.NewRejectedResult(r, entity.StatusRejected, appErr.Message))
	}
	return failedResults
}
