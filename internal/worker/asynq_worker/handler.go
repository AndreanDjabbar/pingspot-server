package asynq_worker

import (
	"context"
	"encoding/json"
	"fmt"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	reportRepo "pingspot/internal/domain/report_service/repository"
	socialRepo "pingspot/internal/domain/social_service/repository"
	redisRepo "pingspot/internal/infrastructure/redis"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"pingspot/internal/task"
	"pingspot/internal/util"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type taskHandler struct {
	DB               *gorm.DB
	ReportRepo       reportRepo.ReportRepository
	NotificationRepo notificationRepo.NotificationRepository
	SocialRepo       socialRepo.FollowRepository
	CacheRepo        redisRepo.CacheRepository
}

func newTaskHandler(db *gorm.DB, cacheRepo redisRepo.CacheRepository, reportRepo reportRepo.ReportRepository, notificationRepo notificationRepo.NotificationRepository, socialRepo socialRepo.FollowRepository) *taskHandler {
	return &taskHandler{
		DB:               db,
		ReportRepo:       reportRepo,
		NotificationRepo: notificationRepo,
		SocialRepo:       socialRepo,
		CacheRepo:        cacheRepo,
	}
}

func sendNotificationNewReportEmails(report model.Report, creator model.User, followers []model.User, reportLink string) error {
	for _, follower := range followers {
		emailData := util.EmailData{
			To:            follower.Email,
			Subject:       fmt.Sprintf("%s Publish a new report!", creator.Username),
			RecipientName: follower.Username,
			EmailType:     util.EmailTypeNewReport,
			TemplateData: map[string]any{
				"ReportTitle": report.ReportTitle,
				"ReportID":    report.ID,
				"UserName":    creator.Username,
				"UserEmail":   creator.Email,
				"ReportLink":  reportLink,
			},
			DisabledNotifications: follower.IsDisableEmailNotification,
			BodyTempate:           getNewReportEmailTemplate(),
		}

		if err := util.SendEmail(emailData); err != nil {
			logger.Error("Failed to send email", zap.String("email", follower.Email), zap.Error(err))
		}
	}
	return nil
}

func sendNotificationReportCommentEmail(report model.Report, comment model.ReportComment, commenter model.User, reportLink string) error {
	emailData := util.EmailData{
		To:            report.User.Email,
		Subject:       fmt.Sprintf("New comment on your report: %s", report.ReportTitle),
		RecipientName: report.User.Username,
		EmailType:     util.EmailTypeReportComment,
		TemplateData: map[string]any{
			"ReportTitle":   report.ReportTitle,
			"ReportID":      report.ID,
			"UserName":      report.User.Username,
			"UserEmail":     report.User.Email,
			"CommenterName": commenter.Username,
			"ReportLink":    reportLink,
		},
		DisabledNotifications: report.User.IsDisableEmailNotification,
		BodyTempate:           getReportCommentEmailTemplate(),
	}

	if err := util.SendEmail(emailData); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func getReportCommentEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>New Comment on Your Report</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			{{.CommenterName}} commented on your report: {{.ReportTitle}}
		</div>
	
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f4f3fb;">
			<tr>
				<td align="center" style="padding: 40px 16px;">
	
					<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; width: 100%; background-color: #ffffff; border-radius: 20px; box-shadow: 0 12px 32px rgba(8, 145, 178, 0.12); overflow: hidden;">
	
						<!-- Header -->
						<tr>
							<td align="center" bgcolor="#0891B2" style="background-color: #0891B2; background-image: linear-gradient(135deg, #22B8D6 0%, #0891B2 50%, #0E7490 100%); padding: 44px 40px 36px;">
								<h1 style="margin: 0; color: #ffffff; font-size: 30px; font-weight: 800; letter-spacing: -0.5px;">
									PingSpot
								</h1>
								<p style="margin: 10px 0 0; color: #cffafe; font-size: 15px; font-weight: 500; letter-spacing: 0.3px;">
									New Comment on Your Report
								</p>
							</td>
						</tr>
	
						<!-- Body -->
						<tr>
							<td style="padding: 44px 40px 20px;">
	
								<!-- Icon badge -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 22px;">
									<tr>
										<td align="center" width="64" height="64" bgcolor="#CFFAFE" style="width: 64px; height: 64px; background-color: #cffafe; border: 1px solid #a5f3fc; border-radius: 50%; font-size: 28px; line-height: 64px;">
											💬
										</td>
									</tr>
								</table>
	
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Hello {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 28px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
										<strong style="color: #0891B2;">{{.CommenterName}}</strong> just left a comment on your report.
								</p>
	
								<!-- Report card -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #cffafe; border: 1px solid #a5f3fc; border-radius: 14px;">
									<tr>
										<td style="padding: 22px 24px;">
											<p style="margin: 0 0 6px; color: #0891B2; font-size: 12px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase;">
												Your Report
											</p>
											<p style="margin: 0; color: #1e1b3a; font-size: 18px; font-weight: 700; line-height: 1.4;">
												📋 {{.ReportTitle}}
											</p>
										</td>
									</tr>
								</table>
	
								
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 36px auto 14px;">
									<tr>
										<td align="center" bgcolor="#0891B2" style="background-color: #0891B2; border-radius: 12px; box-shadow: 0 6px 18px rgba(8, 145, 178, 0.35);">
											<a href="{{.ReportLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												View Comment &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<p style="margin: 0 0 8px; color: #8a8fa8; font-size: 12px; text-align: center; line-height: 1.6;">
									Button not working? Copy this link into your browser:<br>
									<a href="{{.ReportLink}}" style="color: #0891B2; text-decoration: underline; word-break: break-all;">{{.ReportLink}}</a>
								</p>
							</td>
						</tr>
	
						<!-- Why you got this -->
						<tr>
							<td style="padding: 8px 40px 40px;">
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%">
									<tr>
										<td style="border-top: 1px solid #ebe9f7; padding-top: 22px; text-align: center;">
											<p style="margin: 0; color: #64688a; font-size: 14px; line-height: 1.6;">
												🔔 You are receiving this email because there is a new comment on a report you own.<br>
												You can turn off email notifications at any time in your account settings.
											</p>
										</td>
									</tr>
								</table>
							</td>
						</tr>
	
						<!-- Footer -->
						<tr>
							<td align="center" style="background-color: #faf9ff; padding: 28px 40px; border-top: 1px solid #ebe9f7;">
								<p style="margin: 0 0 8px; color: #64688a; font-size: 13px; font-weight: 600;">
									© 2026 PingSpot. All rights reserved.
								</p>
								<p style="margin: 0; color: #9498b3; font-size: 12px; line-height: 1.6;">
									Questions? Contact us at
									<a href="mailto:support@pingspot.com" style="color: #0891B2; text-decoration: none; font-weight: 600;">support@pingspot.com</a>
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

func getNewReportEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>New Report on PingSpot</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			{{.UserName}} just published a report: {{.ReportTitle}}
		</div>
	
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f4f3fb;">
			<tr>
				<td align="center" style="padding: 40px 16px;">
	
					<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; width: 100%; background-color: #ffffff; border-radius: 20px; box-shadow: 0 12px 32px rgba(8, 145, 178, 0.12); overflow: hidden;">
	
						<!-- Header -->
						<tr>
							<td align="center" bgcolor="#0891B2" style="background-color: #0891B2; background-image: linear-gradient(135deg, #22B8D6 0%, #0891B2 50%, #0E7490 100%); padding: 44px 40px 36px;">
								<h1 style="margin: 0; color: #ffffff; font-size: 30px; font-weight: 800; letter-spacing: -0.5px;">
									PingSpot
								</h1>
								<p style="margin: 10px 0 0; color: #cffafe; font-size: 15px; font-weight: 500; letter-spacing: 0.3px;">
									New Report From Someone You Follow
								</p>
							</td>
						</tr>
	
						<!-- Body -->
						<tr>
							<td style="padding: 44px 40px 20px;">
	
								<!-- Icon badge -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 22px;">
									<tr>
										<td align="center" width="64" height="64" bgcolor="#CFFAFE" style="width: 64px; height: 64px; background-color: #cffafe; border: 1px solid #a5f3fc; border-radius: 50%; font-size: 28px; line-height: 64px;">
											📢
										</td>
									</tr>
								</table>
	
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									There's a new report! 👋
								</h2>
								<p style="margin: 0 0 28px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
										<strong style="color: #0891B2;">{{.UserName}}</strong> just published a new report. Check out the details now.
								</p>
	
								<!-- Report card -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #cffafe; border: 1px solid #a5f3fc; border-radius: 14px;">
									<tr>
										<td style="padding: 22px 24px;">
											<p style="margin: 0 0 6px; color: #0891B2; font-size: 12px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase;">
												New Report
											</p>
											<p style="margin: 0 0 10px; color: #1e1b3a; font-size: 18px; font-weight: 700; line-height: 1.4;">
												📋 {{.ReportTitle}}
											</p>
											<p style="margin: 0; color: #64688a; font-size: 14px;">
												Published by <strong style="color: #3b3866;">{{.UserName}}</strong>
											</p>
										</td>
									</tr>
								</table>
	
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 36px auto 14px;">
									<tr>
										<td align="center" bgcolor="#0891B2" style="background-color: #0891B2; border-radius: 12px; box-shadow: 0 6px 18px rgba(8, 145, 178, 0.35);">
											<a href="{{.ReportLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												View Report &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<p style="margin: 0 0 8px; color: #8a8fa8; font-size: 12px; text-align: center; line-height: 1.6;">
									Button not working? Copy this link into your browser:<br>
									<a href="{{.ReportLink}}" style="color: #0891B2; text-decoration: underline; word-break: break-all;">{{.ReportLink}}</a>
								</p>
							</td>
						</tr>
	
						<!-- Why you got this -->
						<tr>
							<td style="padding: 8px 40px 40px;">
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%">
									<tr>
										<td style="border-top: 1px solid #ebe9f7; padding-top: 22px; text-align: center;">
											<p style="margin: 0; color: #64688a; font-size: 14px; line-height: 1.6;">
												🔔 You're receiving this email because you follow <strong>{{.UserName}}</strong> on PingSpot.<br>
												You can turn off email notifications at any time in your account settings.
											</p>
										</td>
									</tr>
								</table>
							</td>
						</tr>
	
						<!-- Footer -->
						<tr>
							<td align="center" style="background-color: #faf9ff; padding: 28px 40px; border-top: 1px solid #ebe9f7;">
								<p style="margin: 0 0 8px; color: #64688a; font-size: 13px; font-weight: 600;">
									© 2026 PingSpot. All rights reserved.
								</p>
								<p style="margin: 0; color: #9498b3; font-size: 12px; line-height: 1.6;">
									Questions? Contact us at
									<a href="mailto:andreanjabar18@gmail.com" style="color: #0891B2; text-decoration: none; font-weight: 600;">andreanjabar18@gmail.com</a>
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

func (h *taskHandler) autoResolveReportHandler(ctx context.Context, t *asynq.Task) error {
	var payload task.UpdateProgressPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	tx := h.DB.Begin()
	report, err := h.ReportRepo.GetByIDTX(ctx, tx, payload.ReportID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("report not found: %w", err)
	}

	if report.ReportStatus == "POTENTIALLY_RESOLVED" {
		if report.PotentiallyResolvedAt == nil {
			tx.Rollback()
			return fmt.Errorf("report %d has POTENTIALLY_RESOLVED status but PotentiallyResolvedAt is nil", report.ID)
		}

		lastUpdate := time.Unix(*report.PotentiallyResolvedAt, 0)
		if time.Since(lastUpdate) >= 20*time.Minute {
			report.ReportStatus = "RESOLVED"
			report.LastUpdatedProgressAt = util.Int64PtrOrNil(time.Now().Unix())
			report.LastUpdatedBy = model.System
			if _, err := h.ReportRepo.UpdateTX(ctx, tx, report); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update report: %w", err)
			}
			tx.Commit()
			logger.Info("Auto resolve report handler success for", zap.Int("report_id", int(report.ID)))
		} else {
			tx.Rollback()
		}
	} else {
		tx.Rollback()
	}

	return nil
}

func (h *taskHandler) CreateNotificationHandler(ctx context.Context, t *asynq.Task) error {
	var payload task.CreateNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	Tx := h.DB.Begin()
	notification := &model.Notification{
		UserID:      payload.UserID,
		Title:       payload.Title,
		Description: payload.Description,
		EntityID:    payload.EntityID,
		EntityType:  &payload.EntityType,
		Category:    payload.Category,
		Type:        payload.Type,
		IsRead:      util.BoolPtrOrNil(false),
	}

	if err := h.NotificationRepo.CreateTX(ctx, Tx, notification); err != nil {
		Tx.Rollback()
		return fmt.Errorf("failed to create notification: %w", err)
	}
	Tx.Commit()

	return nil
}

func (h *taskHandler) SendFollowerReportNotificationHandler(ctx context.Context, t *asynq.Task) error {
	var payload task.SendFollowerReportNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	followers, err := h.SocialRepo.GetFollowersByUserID(ctx, payload.User.ID)
	if err != nil {
		return fmt.Errorf("failed to get followers: %w", err)
	}

	if len(followers) == 0 {
		logger.Info("No followers found for user", zap.Int("user_id", int(payload.User.ID)))
		return nil
	}

	const cooldownDuration = 10 * time.Hour

	var activeFollowerIDs []string
	followerMap := make(map[string]model.User)

	for _, follower := range followers {
		if follower.IsDisableEmailNotification {
			continue
		}

		idStr := fmt.Sprintf("%d", follower.ID)
		activeFollowerIDs = append(activeFollowerIDs, idStr)
		followerMap[idStr] = *follower
	}

	if len(activeFollowerIDs) == 0 {
		logger.Info("All followers have disabled email notifications", zap.Int("user_id", int(payload.User.ID)))
		return nil
	}

	allowedFollowerIDs, err := h.CacheRepo.FilterByCooldown(
		ctx,
		"email:follower_new_report",
		activeFollowerIDs,
		fmt.Sprintf("%d", payload.User.ID),
		cooldownDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to filter followers by cooldown: %w", err)
	}

	if len(allowedFollowerIDs) == 0 {
		logger.Info("No followers allowed to receive notification due to cooldown", zap.Int("user_id", int(payload.User.ID)))
		return nil
	}

	eligibleFollowers := make([]model.User, len(allowedFollowerIDs))
	for i, idStr := range allowedFollowerIDs {
		eligibleFollowers[i] = followerMap[idStr]
	}

	reportLink := fmt.Sprintf("%s/main/reports/%d", util.ClientURL(), payload.Report.ID)

	if err := sendNotificationNewReportEmails(payload.Report, payload.User, eligibleFollowers, reportLink); err != nil {
		logger.Error("Failed to send report notification emails", zap.Error(err))
		return fmt.Errorf("failed to send notification emails: %w", err)
	}

	return nil
}

func (h *taskHandler) SendReportCommentNotificationHandler(ctx context.Context, t *asynq.Task) error {
	var payload task.SendReportCommentNotificationPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.Report.User.IsDisableEmailNotification {
		logger.Info("Report owner has disabled email notifications", zap.Int("user_id", int(payload.Report.UserID)), zap.Int("report_id", int(payload.Report.ID)))
		return nil
	}

	const cooldownDuration = 4 * time.Hour

	allowedIDs, err := h.CacheRepo.FilterByCooldown(
		ctx,
		"email:report_comment",
		[]string{fmt.Sprintf("%d", payload.Report.UserID)},
		fmt.Sprintf("%d", payload.Report.ID),
		cooldownDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to filter comment cooldown: %w", err)
	}

	if len(allowedIDs) == 0 {
		logger.Info("Comment email suppressed due to cooldown", zap.Int("user_id", int(payload.Report.UserID)), zap.Int("report_id", int(payload.Report.ID)))
		return nil
	}

	reportLink := fmt.Sprintf("%s/main/reports/%d", util.ClientURL(), payload.Report.ID)

	if err := sendNotificationReportCommentEmail(payload.Report, payload.Comment, payload.Commenter, reportLink); err != nil {
		logger.Error("Failed to send report comment notification emails", zap.Error(err))
		return fmt.Errorf("failed to send report comment notification emails: %w", err)
	}

	return nil
}
