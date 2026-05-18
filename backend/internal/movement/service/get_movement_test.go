package service

import (
	"context"
	"errors"
	"testing"

	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
)

func TestGetMovementsByItemID_InvalidItemID(t *testing.T) {
	svc := newMovementService(nil, nil)

	_, err := svc.GetMovementsByItemID(context.Background(), 0, nil)
	if err == nil {
		t.Fatal("expected error for item id <= 0, got nil")
	}
	if err.Error() != "invalid item id for history lookup" {
		t.Errorf("expected error 'invalid item id for history lookup', got '%s'", err.Error())
	}
}

func TestGetMovementsByItemID_Success(t *testing.T) {
	expectedMovements := []*movementEntity.Movement{
		{ExternalID: "EXT-001", ItemID: 100, Type: movementEntity.MovementTypeIn, Quantity: 10},
		{ExternalID: "EXT-002", ItemID: 100, Type: movementEntity.MovementTypeOut, Quantity: 5},
	}

	var capturedItemID int
	var capturedPaging *core.Pagination

	mr := &mockMovementRepo{
		getMovementsByItemIDFn: func(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
			capturedItemID = itemId
			capturedPaging = paging
			return expectedMovements, nil
		},
	}

	svc := newMovementService(mr, nil)

	paging := &core.Pagination{Page: 1, Limit: 10}
	result, err := svc.GetMovementsByItemID(context.Background(), 100, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedItemID != 100 {
		t.Errorf("expected item id 100 to be passed to repo, got %d", capturedItemID)
	}
	if capturedPaging != paging {
		t.Errorf("expected paging object to be passed to repo")
	}
	if len(result) != len(expectedMovements) {
		t.Fatalf("expected %d movements, got %d", len(expectedMovements), len(result))
	}
	if result[0].ExternalID != "EXT-001" || result[1].ExternalID != "EXT-002" {
		t.Error("movements data mismatch")
	}
}

func TestGetMovementsByItemID_RepoError(t *testing.T) {
	wantErr := errors.New("database connection failed")

	mr := &mockMovementRepo{
		getMovementsByItemIDFn: func(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
			return nil, wantErr
		},
	}

	svc := newMovementService(mr, nil)

	_, err := svc.GetMovementsByItemID(context.Background(), 100, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error %v, got %v", wantErr, err)
	}
}
