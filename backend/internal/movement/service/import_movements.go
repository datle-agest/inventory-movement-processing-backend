package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"mime/multipart"
	"sync"
)

func (s *service) ImportBatch(ctx context.Context, file *multipart.FileHeader) (entity.ImportBatchResult, error) {
	// validate file check định dạng, dung lượng
	if err := s.validateFile(file); err != nil {
		s.logger.Warnf("[Service][ImportBatch] file validation failed: %v", err)
		return entity.ImportBatchResult{}, err
	}

	// open file
	f, err := file.Open()
	if err != nil {
		s.logger.Errorf("[Service][ImportBatch] failed to open uploaded file: %v", err)
		return entity.ImportBatchResult{}, common.ErrInternal("cannot open uploaded file")
	}
	defer f.Close()

	// parse csv
	validRows, parseFailedRows, err := s.parseCSV(f)
	if err != nil {
		var appErr *common.AppError
		if errors.As(err, &appErr) {
			s.logger.Warnf("[Service][ImportBatch] csv parse logic error: %v", appErr.Message)
			return entity.ImportBatchResult{}, appErr
		}

		s.logger.Errorf("[Service][ImportBatch] failed to parse CSV file: %v", err)
		return entity.ImportBatchResult{}, common.ErrInternal("failed to parse CSV file")
	}

	// group by item_id
	groupedRows := s.groupRowsByItem(validRows)

	// run workers
	resultCh := s.runImportWorkers(ctx, groupedRows)

	totalRows := len(validRows) + len(parseFailedRows)

	// summarize
	result := s.summarizeResults(totalRows, parseFailedRows, resultCh)

	return result, nil
}

// groupRowsByItem - Group rows by item_id for sequential processing per item
func (s *service) groupRowsByItem(rows []entity.CsvMovementRow) map[int32][]entity.CsvMovementRow {
	result := make(map[int32][]entity.CsvMovementRow)
	for _, row := range rows {
		result[row.ItemID] = append(result[row.ItemID], row)
	}
	return result
}

func (s *service) runImportWorkers(ctx context.Context, grouped map[int32][]entity.CsvMovementRow) chan entity.ProcessResult {
	totalRows := 0
	for _, rows := range grouped {
		totalRows += len(rows)
	}
	resultCh := make(chan entity.ProcessResult, totalRows)

	var wg sync.WaitGroup

	// Submit one job per item group
	// Movements of same item processed sequentially
	// Different items processed concurrently
	for _, itemRows := range grouped {

		rows := itemRows
		wg.Add(1)
		s.workerPool.Submit(func() {
			defer wg.Done()
			// sequential within same item
			for _, r := range rows {

				movement := &entity.Movement{
					ExternalID:   r.ExternalID,
					ItemID:       r.ItemID,
					Type:         r.Type,
					Quantity:     r.Quantity,
					MovementTime: r.MovementTime,
					Note:         &r.Note,
				}

				status, err := s.ProcessOne(
					ctx,
					movement,
				)

				res := entity.ProcessResult{
					RowIndex:   r.RowIndex,
					ExternalID: r.ExternalID,
					Status:     status,
				}

				if err != nil {
					res.ErrorReason = err.Error()
				}

				resultCh <- res
			}
		})
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

// summarizeResults
func (s *service) summarizeResults(
	total int,
	parseFailedRows []entity.ProcessResult,
	resultCh chan entity.ProcessResult,
) entity.ImportBatchResult {

	var (
		success   int
		rejected  int
		duplicate int
	)

	failedRows := append([]entity.ProcessResult{}, parseFailedRows...)

	rejected = len(parseFailedRows)

	for r := range resultCh {

		switch r.Status {

		case entity.StatusAccepted:
			success++

		case entity.StatusRejected:
			rejected++
			failedRows = append(failedRows, r)

		case entity.StatusDuplicate:
			duplicate++
			failedRows = append(failedRows, r)
		}
	}

	return entity.ImportBatchResult{
		Total:      total,
		Success:    success,
		Rejected:   rejected,
		Duplicate:  duplicate,
		FailedRows: failedRows,
	}
}
