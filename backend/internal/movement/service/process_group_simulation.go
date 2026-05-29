package service

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
)

func (s *service) simulateMovements(initialStock int32, rows []entity.CsvMovementRow, existingMap map[string]bool) (int32, []*entity.Movement, []entity.ProcessResult) {
	tempStock := initialStock
	var successfulMovements []*entity.Movement
	var results []entity.ProcessResult

	// Keep track of external_ids seen in this batch
	seenInBatch := make(map[string]bool)

	for _, r := range rows {
		// Intra-batch duplicate check - check trong cùng cvs có duplicate k
		if seenInBatch[r.ExternalID] {
			results = append(results, entity.NewRejectedResult(r, entity.StatusDuplicate, "duplicate external_id in the same batch"))
			continue
		}
		seenInBatch[r.ExternalID] = true

		// Database duplicate check
		if existingMap[r.ExternalID] {
			results = append(results, entity.NewRejectedResult(r, entity.StatusDuplicate, "duplicate transaction detected"))
			continue
		}

		// Create movement entity to validate logic
		movement := entity.NewMovementFromCSV(r)
		if err := movement.Validate(); err != nil {
			results = append(results, entity.NewRejectedResult(r, entity.StatusRejected, err.Error()))
			continue
		}

		// Calculate quantity change
		quantityChange, err := calculateQuantityChange(movement)
		if err != nil {
			results = append(results, entity.NewRejectedResult(r, entity.StatusRejected, err.Error()))
			continue
		}

		// Check negative stock business rule
		if tempStock+quantityChange < 0 {
			results = append(results, entity.NewRejectedResult(r, entity.StatusRejected, "insufficient stock"))
			continue
		}

		// Accept this movement
		tempStock += quantityChange
		successfulMovements = append(successfulMovements, movement)
		results = append(results, entity.NewAcceptedResult(r))
	}

	return tempStock, successfulMovements, results
}

// calculateQuantityChange is a pure function - no service dependency needed
func calculateQuantityChange(movement *entity.Movement) (int32, error) {
	switch movement.Type {
	case entity.MovementTypeIn:
		return movement.Quantity, nil
	case entity.MovementTypeOut:
		return -movement.Quantity, nil
	case entity.MovementTypeAdjust:
		return movement.Quantity, nil // ADJUST is relative stock delta
	default:
		return 0, common.NewBadRequestError(common.CodeInvalidMovementType, "invalid movement type")
	}
}
