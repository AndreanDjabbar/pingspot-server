package cron_worker

import (
	"context"
	"fmt"
	"pingspot/internal/domain/report_service/repository"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"pingspot/internal/task"
	"pingspot/internal/util"
	"time"

	"gorm.io/gorm"
)

type cronHandler struct {
	db           *gorm.DB
	reportRepo   repository.ReportRepository
	tasksService task.TaskService
}

func newCronHandler(db *gorm.DB, reportRepo repository.ReportRepository, tasksService task.TaskService) *cronHandler {
	return &cronHandler{
		db:           db,
		reportRepo:   reportRepo,
		tasksService: tasksService,
	}
}

func getAutoResolvedRemainingDay(report model.Report) int {
    potentiallyResolvedAt := time.Unix(*report.PotentiallyResolvedAt, 0)
    autoResolveTime := potentiallyResolvedAt.Add(7 * 24 * time.Hour)
    now := time.Now()
    remaining := autoResolveTime.Sub(now)
    if remaining <= 0 {
        return 0
    }

    daysLeft := int(remaining.Hours() / 24)
    return daysLeft
}

func sendAutoResolvedRemainingDayEmail(to, username, reportTitle, reportLink string, daysRemaining int) error {
	return util.SendEmail(util.EmailData{
		To:            to,
		Subject:       "Pemberitahuan: Laporan Anda Akan Otomatis Ditandai Sebagai Terselesaikan",
		RecipientName: username,
		EmailType:     util.EmailTypeProgressReminder,
		TemplateData: map[string]any{
			"UserName":      username,
			"ReportTitle":   reportTitle,
			"ReportLink":    reportLink,
			"DaysRemaining": daysRemaining,
		},
		BodyTempate: getAutoResolveRemainingDayEmailTemplate(),
	})
}

func getAutoResolveRemainingDayEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Pengingat Progress Laporan</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif; background-color: #f8fafc; line-height: 1.6;">
	<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f8fafc;">
		<tr>
			<td align="center" style="padding: 40px 20px;">
				<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1); overflow: hidden;">
					<tr>
						<td style="background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); padding: 40px 40px 30px; text-align: center;">
							<h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 700; letter-spacing: -0.5px;">
								PingSpot
							</h1>
							<p style="margin: 8px 0 0; color: rgba(255, 255, 255, 0.9); font-size: 16px; font-weight: 400;">
								Pengingat Progress Laporan
							</p>
						</td>
					</tr>
					<tr>
						<td style="padding: 50px 40px;">
							<h2 style="margin: 0 0 20px; color: #1e293b; font-size: 24px; font-weight: 600; text-align: center;">
								Halo {{.UserName}}! 👋
							</h2>
							<p style="margin: 0 0 25px; color: #475569; font-size: 16px; text-align: center; line-height: 1.7;">
								Laporan Anda berstatus <strong style="color: #f59e0b;">Dalam Peninjauan</strong> dan memerlukan pembaruan segera!
							</p>
							<div style="margin: 30px 0; padding: 25px; background-color: #fef3c7; border-radius: 12px; border-left: 4px solid #f59e0b;">
								<p style="margin: 0 0 15px; color: #92400e; font-size: 16px; font-weight: 600;">
									📋 {{.ReportTitle}}
								</p>
								<div style="margin: 20px 0; padding: 15px; background-color: #ffffff; border-radius: 8px; border: 2px solid #f59e0b;">
									<p style="margin: 0; color: #78350f; font-size: 18px; font-weight: 700; text-align: center;">
										⏰ {{.DaysRemaining}} Hari Tersisa
									</p>
								</div>
								<p style="margin: 15px 0 0; color: #78350f; font-size: 14px; line-height: 1.6;">
									<strong>⚠️ Tindakan Diperlukan:</strong><br>
									Harap unggah bukti progress laporan Anda sebelum batas waktu berakhir. Jika tidak ada pembaruan dalam {{.DaysRemaining}} hari ke depan, laporan akan <strong>otomatis ditandai sebagai Terselesaikan</strong> oleh sistem.
								</p>
							</div>
							<div style="margin: 25px 0; padding: 20px; background-color: #eff6ff; border-radius: 12px; border-left: 4px solid #3b82f6;">
								<p style="margin: 0 0 10px; color: #1e40af; font-size: 14px; font-weight: 600;">
									💡 Apa yang perlu Anda lakukan?
								</p>
								<ul style="margin: 0; padding-left: 20px; color: #1e40af; font-size: 14px; line-height: 1.8;">
									<li>Unggah foto atau bukti progress terbaru</li>
									<li>Berikan update status penanganan</li>
									<li>Tambahkan keterangan jika diperlukan</li>
								</ul>
							</div>
							<div style="text-align: center; margin: 35px 0;">
								<a href="{{.ReportLink}}" 
								   style="display: inline-block; background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: #ffffff; text-decoration: none; padding: 16px 32px; border-radius: 50px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 15px rgba(245, 158, 11, 0.4); transition: all 0.3s ease; text-align: center; min-width: 200px;">
									Perbarui Progress Sekarang
								</a>
							</div>
							<div style="margin: 30px 0; text-align: center;">
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.6;">
									Jika Anda memiliki pertanyaan atau kendala, jangan ragu untuk menghubungi kami.
								</p>
							</div>
						</td>
					</tr>
					<tr>
						<td style="background-color: #f8fafc; padding: 30px 40px; text-align: center; border-top: 1px solid #e2e8f0;">
							<p style="margin: 0 0 10px; color: #64748b; font-size: 14px;">
								© 2025 PingSpot. Hak cipta dilindungi undang-undang.
							</p>
							<p style="margin: 0; color: #94a3b8; font-size: 12px;">
								Pertanyaan? Hubungi kami di
								<a href="mailto:support@pingspot.com" style="color: #f59e0b; text-decoration: none; font-weight: 500;">
									support@pingspot.com
								</a>
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>`
}


func (h *cronHandler) checkPotentiallyResolvedReport() error {
	logger.Info("Executing CheckPotentiallyResolvedReport cron job")
	ctx := context.Background()
	reports, err := h.reportRepo.GetByReportStatus(ctx, "POTENTIALLY_RESOLVED")
	if err != nil {
		return fmt.Errorf("failed to get potentially resolved reports: %w", err)
	}

	for _, report := range *reports {
		if err := h.tasksService.AutoResolveReportTask(report.ID); err != nil {
			logger.Error(fmt.Sprintf("failed to enqueue auto resolve report task for report ID %d: %v", report.ID, err))
			continue
		}

		if report.User.Email == "" || report.User.Username == "" {
			logger.Error(fmt.Sprintf("user data not properly loaded for report ID %d, skipping email", report.ID))
			continue
		}
		remainingDay := getAutoResolvedRemainingDay(report)
		reportLink := fmt.Sprintf("%s/main/report/%d", util.ClientURL(), report.ID)
		go sendAutoResolvedRemainingDayEmail(report.User.Email, report.User.Username, report.ReportTitle, reportLink, remainingDay)
	}
	return nil
}

func (h *cronHandler) expireOldReports() error {
	logger.Info("Executing ExpireOldReports cron job")
	ctx := context.Background()

	reports, err := h.reportRepo.GetByReportStatus(ctx, "WAITING", "ON_PROGRESS")
	if err != nil {
		return fmt.Errorf("failed to get reports for expiration check: %w", err)
	}

	now := time.Now().Unix()
	threshold := now - (30 * 24 * 60 * 60)

	for _, report := range *reports {
		if report.LastUpdatedProgressAt != nil && report.LastUpdatedBy != "" {
			if *report.LastUpdatedProgressAt <= threshold && report.LastUpdatedBy == "SYSTEM" {
				tx := h.db.Begin()

				report.ReportStatus = "EXPIRED"
				report.LastUpdatedBy = "SYSTEM"
				report.UpdatedAt = time.Now().Unix()

				if _, err := h.reportRepo.UpdateTX(ctx, tx, &report); err != nil {
					tx.Rollback()
					logger.Error(fmt.Sprintf("Failed to expire report ID %d: %v", report.ID, err))
					continue
				}

				if err := tx.Commit().Error; err != nil {
					logger.Error(fmt.Sprintf("Failed to commit transaction for report ID %d: %v", report.ID, err))
					continue
				}

				logger.Info(fmt.Sprintf("Report ID %d has been marked as EXPIRED", report.ID))
			}
		}
	}
	return nil
}

func (h *cronHandler) hardDeleteOldReports() error {
	logger.Info("Executing HardDeleteOldReports cron job")
	ctx := context.Background()

	deletedReports, err := h.reportRepo.GetByIsDeleted(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to get soft-deleted reports: %w", err)
	}

	now := time.Now().Unix()
	threshold := util.Int64PtrOrNil(now - (30 * 24 * 60 * 60))

	for _, deletedReport := range deletedReports {
		if deletedReport.DeletedAt != nil && threshold != nil && *deletedReport.DeletedAt <= *threshold {
			tx := h.db.Begin()

			if _, err := h.reportRepo.DeleteTX(ctx, tx, deletedReport); err != nil {
				tx.Rollback()
				logger.Error(fmt.Sprintf("Failed to hard delete report ID %d: %v", deletedReport.ID, err))
				continue
			}

			if err := tx.Commit().Error; err != nil {
				logger.Error(fmt.Sprintf("Failed to commit transaction for hard deleting report ID %d: %v", deletedReport.ID, err))
				continue
			}
			logger.Info(fmt.Sprintf("Report ID %d has been hard deleted", deletedReport.ID))
		}
	}
	return nil
}
