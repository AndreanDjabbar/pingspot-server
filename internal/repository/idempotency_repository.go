package repository

import (
	"context"
	"errors"
	"pingspot/internal/model"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrIdempotencyRecordConflict = errors.New("idempotency record already exists")

type IdempotencyRepository interface {
	Get(ctx context.Context, userID uint, idempotencyKey, method, route string) (*model.IdempotencyRecord, error)
	Create(ctx context.Context, record *model.IdempotencyRecord) error
	Complete(ctx context.Context, idempotencyID uint, responseStatus int, body []byte) error
	Delete(ctx context.Context, idempotencyID uint) error
}

type idempotencyRepository struct {
	db *gorm.DB
}

func NewIdempotencyRepository(db *gorm.DB) IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) Get(ctx context.Context, userID uint, idempotencyKey, method, route string) (*model.IdempotencyRecord, error) {
	var record model.IdempotencyRecord
	err := r.db.WithContext(ctx).Where("user_id = ? AND idempotency_key = ? AND method = ? AND route = ?", userID, idempotencyKey, method, route).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *idempotencyRepository) Create(ctx context.Context, record *model.IdempotencyRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		var pgError *pgconn.PgError
		if errors.Is(err, gorm.ErrDuplicatedKey) || (errors.As(err, &pgError) && pgError.Code == "23505") {
			return ErrIdempotencyRecordConflict
		}
		return err
	}
	return nil
}

func (r *idempotencyRepository) Complete(ctx context.Context, idempotencyID uint, responseStatus int, body []byte) error {
	now := time.Now()
	updates := map[string]any{
		"status":          "COMPLETED",
		"response_status": responseStatus,
		"response_body":   body,
		"completed_at":    now,
	}
	if err := r.db.WithContext(ctx).Model(&model.IdempotencyRecord{}).Where("id = ?", idempotencyID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (r *idempotencyRepository) Delete(ctx context.Context, idempotencyID uint) error {
	return r.db.WithContext(ctx).Delete(&model.IdempotencyRecord{}, idempotencyID).Error
}