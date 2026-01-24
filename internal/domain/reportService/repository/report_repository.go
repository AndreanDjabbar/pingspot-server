package repository

import (
	"context"
	"pingspot/internal/domain/reportService/dto"
	"pingspot/internal/model"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type ReportRepository interface {
	Create(ctx context.Context, report *model.Report, tx *gorm.DB) error
	UpdateTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error)
	GetByID(ctx context.Context, reportID uint) (*model.Report, error)
	GetByIDTX(ctx context.Context, tx *gorm.DB, reportID uint) (*model.Report, error)
	Get(ctx context.Context) (*[]model.Report, error)
	GetByReportStatus(ctx context.Context, status ...string) (*[]model.Report, error)
	GetByReportStatusCount(ctx context.Context, status ...string) (map[string]int64, error)
	GetByIDIsDeleted(ctx context.Context, reportID uint, isDeleted bool) (*model.Report, error)
	GetByIsDeleted(ctx context.Context, isDeleted bool) ([]*model.Report, error)
	GetByIsDeletedPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance, isDeleted bool) (*[]model.Report, error)
	GetPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance) (*[]model.Report, error)
	GetByReportTypeCount(ctx context.Context) (*dto.TotalReportCount, error)
	GetMonthlyReportCount(ctx context.Context) (map[string]int64, error)
	FullTextSearchReport(ctx context.Context, searchQuery string, limit int) (*[]model.Report, error)
	FullTextSearchReportPaginated(ctx context.Context, searchQuery string, limit int, cursorID uint) (*[]model.Report, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetByReportTypeCount(ctx context.Context) (*dto.TotalReportCount, error) {
	var grouped []struct {
		ReportType string
		Total      int64
	}

	if err := r.db.WithContext(ctx).Model(&model.Report{}).
		Select("report_type, COUNT(*) as total").
		Group("report_type").
		Scan(&grouped).Error; err != nil {
		return nil, err
	}

	var result dto.TotalReportCount
	for _, g := range grouped {
		switch g.ReportType {
		case "INFRASTRUCTURE":
			result.TotalInfrastructureReports = g.Total
		case "ENVIRONMENT":
			result.TotalEnvironmentReports = g.Total
		case "SAFETY":
			result.TotalSafetyReports = g.Total
		case "TRAFFIC":
			result.TotalTrafficReports = g.Total
		case "PUBLIC_FACILITY":
			result.TotalPublicFacilityReports = g.Total
		case "WASTE":
			result.TotalWasteReports = g.Total
		case "WATER":
			result.TotalWaterReports = g.Total
		case "ELECTRICITY":
			result.TotalElectricityReports = g.Total
		case "HEALTH":
			result.TotalHealthReports = g.Total
		case "SOCIAL":
			result.TotalSocialReports = g.Total
		case "EDUCATION":
			result.TotalEducationReports = g.Total
		case "ADMINISTRATIVE":
			result.TotalAdministrativeReports = g.Total
		case "DISASTER":
			result.TotalDisasterReports = g.Total
		case "OTHER":
			result.TotalOtherReports = g.Total
		}
		result.TotalReports += g.Total
	}

	return &result, nil
}

func (r *reportRepository) GetByReportStatusCount(ctx context.Context, status ...string) (map[string]int64, error) {
	var results []struct {
		ReportStatus string
		Count        int64
	}
	if err := r.db.WithContext(ctx).Model(&model.Report{}).
		Select("report_status, COUNT(*) AS count").
		Where("report_status IN ?", status).
		Group("report_status").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	statusCounts := make(map[string]int64)
	for _, result := range results {
		statusCounts[result.ReportStatus] = result.Count
	}
	return statusCounts, nil
}

func (r *reportRepository) FullTextSearchReport(ctx context.Context, searchQuery string, limit int) (*[]model.Report, error) {
	var reports []model.Report

	if strings.TrimSpace(searchQuery) == "" {
		return &reports, nil
	}

	searchQuery = strings.ToLower(searchQuery)
	searchQuery = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(searchQuery, "")
	searchQuery = strings.TrimSpace(searchQuery)
	searchQuery = regexp.MustCompile(`\s+`).ReplaceAllString(searchQuery, " & ")
	searchQuery += ":*"

	err := r.db.WithContext(ctx).Raw(`
		SELECT *
		FROM reports
		WHERE search_vector @@ to_tsquery('simple', ?)
		ORDER BY ts_rank(search_vector, to_tsquery('simple', ?)) DESC
		LIMIT ?
	`, searchQuery, searchQuery, limit).Scan(&reports).Error

	return &reports, err
}

func (r *reportRepository) GetMonthlyReportCount(
	ctx context.Context,
) (map[string]int64, error) {

	var results []struct {
		Month string
		Count int64
	}

	err := r.db.WithContext(ctx).
		Model(&model.Report{}).
		Select("to_char(to_timestamp(created_at), 'YYYY-MM') AS month, COUNT(*) AS count").
		Group("month").
		Order("month DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	monthlyCounts := make(map[string]int64)
	for _, result := range results {
		monthlyCounts[result.Month] = result.Count
	}

	return monthlyCounts, nil
}

func (r *reportRepository) FullTextSearchReportPaginated(ctx context.Context, searchQuery string, limit int, cursorID uint) (*[]model.Report, error) {
	var reports []model.Report

	if strings.TrimSpace(searchQuery) == "" {
		return &reports, nil
	}

	searchQuery = strings.ToLower(searchQuery)
	searchQuery = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(searchQuery, "")
	searchQuery = strings.TrimSpace(searchQuery)
	searchQuery = regexp.MustCompile(`\s+`).ReplaceAllString(searchQuery, " & ")
	searchQuery += ":*"

	query := `
		SELECT *
		FROM reports
		WHERE search_vector @@ to_tsquery('simple', ?)
	`
	args := []any{searchQuery}

	if cursorID != 0 {
		query += " AND id > ?"
		args = append(args, cursorID)
	}

	query += `
		ORDER BY ts_rank(search_vector, to_tsquery('simple', ?)) DESC, id ASC
		LIMIT ?
	`
	args = append(args, searchQuery, limit)

	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&reports).Error

	return &reports, err
}

func (r *reportRepository) GetByReportStatus(ctx context.Context, status ...string) (*[]model.Report, error) {
	var reports []model.Report
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Profile").
		Where("report_status IN ?", status).
		Find(&reports).Error; err != nil {
		return nil, err
	}
	return &reports, nil
}

func (r *reportRepository) Create(ctx context.Context, report *model.Report, tx *gorm.DB) error {
	if tx != nil {
		return tx.WithContext(ctx).Create(report).Error
	}
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *reportRepository) Get(ctx context.Context) (*[]model.Report, error) {
	var reports []model.Report
	if err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("reports.created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
	}
	return &reports, nil
}

func (r *reportRepository) GetPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance) (*[]model.Report, error) {
	var reportIDs []int64
	var reports []model.Report

	subQuery := r.db.WithContext(ctx).Table("reports")

	if reportType != "" && reportType != "all" {
		subQuery = subQuery.Where("reports.report_type = ?", reportType)
	}

	if status != "" && status != "all" {
		subQuery = subQuery.Where("reports.report_status = ?", status)
	}

	if distance.Distance != "" && distance.Distance != "all" {
		straightDistance := 0
		switch distance.Distance {
		case "1000":
			straightDistance = 1000
		case "5000":
			straightDistance = 5000
		case "10000":
			straightDistance = 10000
		}
		if straightDistance > 0 {
			subQuery = subQuery.
				Joins("JOIN report_locations ON report_locations.report_id = reports.id").
				Where(`
					ST_DWithin(
						report_locations.geometry::geography,
						ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
						?
					)
				`, distance.Lng, distance.Lat, straightDistance)
		}
	}

	if hasProgress != "" && hasProgress != "all" {
		if hasProgress == "true" {
			subQuery = subQuery.Where("reports.has_progress = ?", true)
		} else if hasProgress == "false" {
			subQuery = subQuery.Where("reports.has_progress = ?", false)
		}
	}

	switch sortBy {
	case "latest":
		subQuery = subQuery.Order("reports.id DESC")
	case "oldest":
		subQuery = subQuery.Order("reports.id ASC")
	case "most_liked":
		subQuery = subQuery.
			Joins("LEFT JOIN report_reactions ON reports.id = report_reactions.report_id AND report_reactions.type = 'LIKE'").
			Group("reports.id").
			Order("COUNT(report_reactions.id) DESC")
	case "least_liked":
		subQuery = subQuery.
			Joins("LEFT JOIN report_reactions ON reports.id = report_reactions.report_id AND report_reactions.type = 'LIKE'").
			Group("reports.id").
			Order("COUNT(report_reactions.id) ASC")
	default:
		subQuery = subQuery.Order("reports.id DESC")
	}

	if cursorID != 0 {
		if sortBy == "oldest" || sortBy == "least_liked" {
			subQuery = subQuery.Where("reports.id > ?", cursorID)
		} else {
			subQuery = subQuery.Where("reports.id < ?", cursorID)
		}
	}

	subQuery = subQuery.Limit(int(limit))

	if err := subQuery.Select("reports.id").Pluck("id", &reportIDs).Error; err != nil {
		return nil, err
	}

	if len(reportIDs) == 0 {
		return &reports, nil
	}

	query := r.db.
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("id IN ?", reportIDs)

	switch sortBy {
	case "oldest", "least_liked":
		query = query.Order("id ASC")
	default:
		query = query.Order("id DESC")
	}

	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}

	return &reports, nil
}

func (r *reportRepository) GetByIsDeletedPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance, isDeleted bool) (*[]model.Report, error) {
	var reportIDs []int64
	var reports []model.Report

	subQuery := r.db.WithContext(ctx).Table("reports")

	if reportType != "" && reportType != "all" {
		subQuery = subQuery.Where("reports.report_type = ?", reportType)
	}

	if status != "" && status != "all" {
		subQuery = subQuery.Where("reports.report_status = ?", status)
	}

	if distance.Distance != "" && distance.Distance != "all" {
		straightDistance := 0
		switch distance.Distance {
		case "1000":
			straightDistance = 1000
		case "5000":
			straightDistance = 5000
		case "10000":
			straightDistance = 10000
		}
		if straightDistance > 0 {
			subQuery = subQuery.
				Joins("JOIN report_locations ON report_locations.report_id = reports.id").
				Where(`
					ST_DWithin(
						report_locations.geometry::geography,
						ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
						?
					)
				`, distance.Lng, distance.Lat, straightDistance)
		}
	}

	if hasProgress != "" && hasProgress != "all" {
		if hasProgress == "true" {
			subQuery = subQuery.Where("reports.has_progress = ?", true)
		} else if hasProgress == "false" {
			subQuery = subQuery.Where("reports.has_progress = ?", false)
		}
	}

	switch sortBy {
	case "latest":
		subQuery = subQuery.Order("reports.id DESC")
	case "oldest":
		subQuery = subQuery.Order("reports.id ASC")
	case "most_liked":
		subQuery = subQuery.
			Joins("LEFT JOIN report_reactions ON reports.id = report_reactions.report_id AND report_reactions.type = 'LIKE'").
			Group("reports.id").
			Order("COUNT(report_reactions.id) DESC")
	case "least_liked":
		subQuery = subQuery.
			Joins("LEFT JOIN report_reactions ON reports.id = report_reactions.report_id AND report_reactions.type = 'LIKE'").
			Group("reports.id").
			Order("COUNT(report_reactions.id) ASC")
	default:
		subQuery = subQuery.Order("reports.id DESC")
	}

	if cursorID != 0 {
		if sortBy == "oldest" || sortBy == "least_liked" {
			subQuery = subQuery.Where("reports.id > ?", cursorID)
		} else {
			subQuery = subQuery.Where("reports.id < ?", cursorID)
		}
	}

	subQuery = subQuery.Limit(int(limit))

	if err := subQuery.Select("reports.id").Pluck("id", &reportIDs).Error; err != nil {
		return nil, err
	}

	if len(reportIDs) == 0 {
		return &reports, nil
	}

	query := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("id IN ?", reportIDs).
		Where("is_deleted = ?", isDeleted)

	switch sortBy {
	case "oldest", "least_liked":
		query = query.Order("id ASC")
	default:
		query = query.Order("id DESC")
	}

	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}

	return &reports, nil
}

func (r *reportRepository) UpdateTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error) {
	if err := tx.WithContext(ctx).Save(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (r *reportRepository) DeleteTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error) {
	if err := tx.WithContext(ctx).Delete(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (r *reportRepository) GetByID(ctx context.Context, reportID uint) (*model.Report, error) {
	var report model.Report

	if err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&report, "reports.id = ?", reportID).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *reportRepository) GetByIDIsDeleted(ctx context.Context, reportID uint, isDeleted bool) (*model.Report, error) {
	var report model.Report

	if err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("is_deleted = ?", isDeleted).
		First(&report, "reports.id = ?", reportID).
		Error; err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *reportRepository) GetByIsDeleted(ctx context.Context, isDeleted bool) ([]*model.Report, error) {
	var report []*model.Report

	if err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("is_deleted = ?", isDeleted).
		Find(&report).
		Error; err != nil {
		return nil, err
	}

	return report, nil
}

func (r *reportRepository) GetByIDTX(ctx context.Context, tx *gorm.DB, reportID uint) (*model.Report, error) {
	var report model.Report
	if err := tx.WithContext(ctx).
		Preload("User.Profile").
		Preload("ReportLocation").
		Preload("ReportImages").
		Preload("ReportReactions").
		Preload("ReportVotes").
		Preload("ReportProgress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&report, "reports.id = ?", reportID).Error; err != nil {
		return nil, err
	}
	return &report, nil
}
