package repository

import (
	"context"
	"pingspot/internal/domain/report_service/dto"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportSavedRepository interface {
	CreateTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error
	GetByUserIDAndReportID(ctx context.Context, userID uint, reportID uint) (*model.ReportSaved, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error
	GetByUserIDIsDeletedPaginated(ctx context.Context, userID uint, isDeleted bool, limit int, cursorID *string) (*[]dto.GetSavedReports, error)
}

type reportSavedRepository struct {
	db *gorm.DB
}

func NewReportSavedRepository(db *gorm.DB) ReportSavedRepository {
	return &reportSavedRepository{db: db}
}

func (r *reportSavedRepository) CreateTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error {
	if err := tx.WithContext(ctx).Create(reportSaved).Error; err != nil {
		return err
	}
	return nil
}

func (r *reportSavedRepository) GetByUserIDAndReportID(ctx context.Context, userID uint, reportID uint) (*model.ReportSaved, error) {
	var reportSaved model.ReportSaved
	if err := r.db.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&reportSaved).Error; err != nil {
		return nil, err
	}
	return &reportSaved, nil
}

func (r *reportSavedRepository) DeleteTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error {
	if err := tx.WithContext(ctx).Delete(reportSaved).Error; err != nil {
		return err
	}
	return nil
}

func (r *reportSavedRepository) GetByUserIDIsDeletedPaginated(
	ctx context.Context,
	userID uint,
	isDeleted bool,
	limit int,
	cursorID *string,
) (*[]dto.GetSavedReports, error) {
	var reportSavedIDs []uint
	var savedReports []dto.GetSavedReports

	subQuery := r.db.WithContext(ctx).
		Table("report_saveds").
		Joins("JOIN reports ON reports.id = report_saveds.report_id").
		Where("report_saveds.user_id = ?", userID).
		Where("reports.is_deleted = ?", isDeleted).
		Order("report_saveds.created_at DESC, report_saveds.id DESC")

	if cursorID != nil && *cursorID != "" {
		subQuery = subQuery.Where(`
			(report_saveds.created_at, report_saveds.id) < (
				SELECT created_at, id FROM report_saveds WHERE id = ?
			)
		`, *cursorID)
	}

	subQuery = subQuery.Limit(limit)

	if err := subQuery.Select("report_saveds.id").Pluck("id", &reportSavedIDs).Error; err != nil {
		return nil, err
	}

	if len(reportSavedIDs) == 0 {
		return &savedReports, nil
	}

	query := r.db.WithContext(ctx).
		Table("report_saveds rs").
		Select(`
			rs.id AS report_saved_id,
			r.id AS report_id,
			r.user_id AS user_id,
			r.report_title AS report_title,
			r.report_type AS report_type,
			r.report_description AS report_description,
			r.report_status AS report_status,
			rl.latitude AS report_latitude,
			rl.longitude AS report_longitude,
			rl.map_zoom AS report_map_zoom,
			rl.country AS report_country,
			rl.country_code AS report_country_code,
			rl.state AS report_state
		`).
		Joins("JOIN reports r ON r.id = rs.report_id").
		Joins("JOIN report_locations rl ON rl.report_id = r.id").
		Where("rs.id IN ?", reportSavedIDs)

	if err := query.Scan(&savedReports).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]dto.GetSavedReports, len(savedReports))
	for _, sr := range savedReports {
		byID[sr.ReportSavedID] = sr
	}
	ordered := make([]dto.GetSavedReports, 0, len(savedReports))
	for _, id := range reportSavedIDs {
		if sr, ok := byID[id]; ok {
			ordered = append(ordered, sr)
		}
	}

	return &ordered, nil
}