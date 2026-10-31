package service

import (
	"context"
	"errors"
	"fmt"
	"pingspot/internal/domain/report_service/dto"
	reportRepository "pingspot/internal/domain/report_service/repository"
	userRepository "pingspot/internal/domain/user_service/repository"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"pingspot/internal/task"
	"pingspot/internal/util"
	"regexp"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getMajorityVote(resolvedVote, onProgressVote int64) *string {
	if resolvedVote >= onProgressVote {
		vote := "RESOLVED"
		return &vote
	}
	if onProgressVote >= resolvedVote {
		vote := "ON_PROGRESS"
		return &vote
	}
	return nil
}

func getVoteTypeOrder(voteCount map[model.ReportStatus]int64) []struct {
	Type  model.ReportStatus
	Count int
} {
	votes := []struct {
		Type  model.ReportStatus
		Count int
	}{
		{model.RESOLVED, int(voteCount[model.RESOLVED])},
		{model.ON_PROGRESS, int(voteCount[model.ON_PROGRESS])},
	}

	sort.Slice(votes, func(i, j int) bool {
		return votes[i].Count > votes[j].Count
	})

	return votes
}

func sendPotentiallyResolvedReportEmail(to, username, reportTitle, reportLink string, daysRemaining int, disabledNotifications bool) error {
	return util.SendEmail(util.EmailData{
		To:            to,
		Subject:       "Reminder: Update Progress on Your Report",
		RecipientName: username,
		EmailType:     util.EmailTypeProgressReminder,
		TemplateData: map[string]any{
			"ReportTitle":   reportTitle,
			"ReportLink":    reportLink,
			"DaysRemaining": daysRemaining,
		},
		DisabledNotifications: disabledNotifications,
		BodyTempate:           getProgressReminderEmailTemplate(),
	})
}

func getProgressReminderEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Report Progress Reminder</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			Your report "{{.ReportTitle}}" needs an update. Upload proof of progress before the period ends.
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
									Report Progress Reminder
								</p>
							</td>
						</tr>
	
						<!-- Body -->
						<tr>
							<td style="padding: 44px 40px 20px;">
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Hello {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 28px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									Your report is still marked as
									<span style="display: inline-block; padding: 2px 10px; background-color: #fff4d6; color: #92400e; border-radius: 999px; font-size: 14px; font-weight: 600;">Under Review</span>
									and is waiting for an update from you.
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
	
								<!-- Deadline notice -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="margin-top: 18px; background-color: #fffbeb; border-left: 4px solid #f59e0b; border-radius: 10px;">
									<tr>
										<td style="padding: 18px 20px;">
											<p style="margin: 0 0 6px; color: #92400e; font-size: 15px; font-weight: 700;">
												⏰ Time remaining: 7 weeks
											</p>
											<p style="margin: 0; color: #78350f; font-size: 14px; line-height: 1.6;">
												Upload proof of progress before this period ends. If there is no update, the report will be
												<strong>automatically marked as Resolved</strong>.
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
												Update Progress Now &rarr;
											</a>
										</td>
									</tr>
								</table>
							</td>
						</tr>
	
						<!-- Help -->
						<tr>
							<td style="padding: 8px 40px 40px;">
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%">
									<tr>
										<td style="border-top: 1px solid #ebe9f7; padding-top: 22px; text-align: center;">
											<p style="margin: 0; color: #64688a; font-size: 14px; line-height: 1.6;">
												If you have any questions, feel free to contact us.
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
									Questions? Contact us by email at
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

func convertToDTO(c *model.ReportComment, u *model.User) *dto.Comment {
	comment := &dto.Comment{
		CommentID: c.ID.Hex(),
		ReportID:  c.ReportID,
		UserInformation: &dto.CommentUserInformation{
			UserID:                     u.ID,
			Bio:                        u.Profile.Bio,
			ProfilePicture:             u.Profile.ProfilePicture,
			Gender:                     u.Profile.Gender,
			Birthday:                   u.Profile.Birthday,
			Username:                   u.Username,
			Email:                      u.Email,
			FullName:                   u.FullName,
			IsDefaultUsername:          u.IsDefaultUsername,
			IsDisableEmailNotification: u.IsDisableEmailNotification,
		},
		Mentions: c.Mentions,
		Content:  c.Content,
		ParentCommentID: func() *string {
			if c.ParentCommentID != nil {
				id := c.ParentCommentID.Hex()
				return &id
			}
			return nil
		}(),
		ThreadRootID: func() *string {
			if c.ThreadRootID != nil {
				id := c.ThreadRootID.Hex()
				return &id
			}
			return nil
		}(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: func() *int64 {
			if c.UpdatedAt != nil {
				t := c.UpdatedAt
				return t
			}
			return nil
		}(),
		TotalReplies: 0,
	}

	if c.Media != nil {
		comment.Media = &model.CommentMedia{
			URL:    c.Media.URL,
			Type:   model.CommentMediaType(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return comment
}

func convertToReplyDTO(c *model.ReportComment, u *model.User, mentions []model.Mention) *dto.CommentReply {
	reply := &dto.CommentReply{
		CommentID: c.ID.Hex(),
		ReportID:  c.ReportID,
		UserInformation: &dto.CommentUserInformation{
			UserID:                     u.ID,
			Bio:                        u.Profile.Bio,
			ProfilePicture:             u.Profile.ProfilePicture,
			Gender:                     u.Profile.Gender,
			Birthday:                   u.Profile.Birthday,
			Username:                   u.Username,
			Email:                      u.Email,
			FullName:                   u.FullName,
			IsDefaultUsername:          u.IsDefaultUsername,
			IsDisableEmailNotification: u.IsDisableEmailNotification,
		},
		Mentions: mentions,
		Content:  c.Content,
		ParentCommentID: func() *string {
			if c.ParentCommentID != nil {
				id := c.ParentCommentID.Hex()
				return &id
			}
			return nil
		}(),
		ThreadRootID: func() *string {
			if c.ThreadRootID != nil {
				id := c.ThreadRootID.Hex()
				return &id
			}
			return nil
		}(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: func() *int64 {
			if c.UpdatedAt != nil {
				t := c.UpdatedAt
				return t
			}
			return nil
		}(),
	}

	if c.Media != nil {
		reply.Media = &model.CommentMedia{
			URL:    c.Media.URL,
			Type:   model.CommentMediaType(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return reply
}

func convertRootCommentsToDTO(comments []*model.ReportComment, users map[uint]*model.User, replyCounts map[string]int64, mentionsMap map[string][]model.Mention) []*dto.Comment {
	rootComments := make([]*dto.Comment, 0)

	for _, comment := range comments {
		user := users[comment.UserID]
		if user == nil {
			continue
		}
		comment.Mentions = mentionsMap[comment.ID.Hex()]
		commentDTO := convertToDTO(comment, user)

		commentID := comment.ID.Hex()
		if count, exists := replyCounts[commentID]; exists {
			commentDTO.TotalReplies = count
		}

		rootComments = append(rootComments, commentDTO)
	}

	sort.Slice(rootComments, func(i, j int) bool {
		return rootComments[i].CreatedAt < rootComments[j].CreatedAt
	})

	return rootComments
}

func convertRepliesToDTO(comments []*model.ReportComment, users map[uint]*model.User, parentsComment map[string]*model.ReportComment, commentMentionsMap map[string][]uint) []*dto.CommentReply {
	replies := make([]*dto.CommentReply, 0, len(comments))

	for _, comment := range comments {
		user := users[comment.UserID]
		if user == nil {
			continue
		}

		mentions := make([]model.Mention, 0)
		for _, mentionID := range commentMentionsMap[comment.ID.Hex()] {
			if user, exists := users[mentionID]; exists {
				mentions = append(mentions, model.Mention{
					UserID:   user.ID,
					Username: user.Username,
				})
			}
		}

		replyDTO := convertToReplyDTO(comment, user, mentions)
		if comment.ParentCommentID != nil {
			parentID := comment.ParentCommentID.Hex()
			if parentComment, exists := parentsComment[parentID]; exists {
				if parentUser, userExists := users[parentComment.UserID]; userExists {
					if replyDTO.UserInformation.UserID != parentUser.ID {
						replyDTO.ReplyTo = parentUser
					}
				}
			}
		}

		replies = append(replies, replyDTO)
	}

	sort.Slice(replies, func(i, j int) bool {
		return replies[i].CreatedAt < replies[j].CreatedAt
	})

	return replies
}

type ReportService struct {
	postgreDB          *gorm.DB
	mongoDB            *mongo.Client
	reportRepo         reportRepository.ReportRepository
	reportLocationRepo reportRepository.ReportLocationRepository
	reportImageRepo    reportRepository.ReportImageRepository
	reportReactionRepo reportRepository.ReportReactionRepository
	reportVoteRepo     reportRepository.ReportVoteRepository
	reportProgressRepo reportRepository.ReportProgressRepository
	tasksService       task.TaskService
	userRepo           userRepository.UserRepository
	userProfileRepo    userRepository.UserProfileRepository
	reportCommentRepo  reportRepository.ReportCommentRepository
	reportSavedRepo    reportRepository.ReportSavedRepository
}

func NewreportService(
	postgreDB *gorm.DB,
	mongoDB *mongo.Client,
	reportRepo reportRepository.ReportRepository,
	locationRepo reportRepository.ReportLocationRepository,
	reportReaction reportRepository.ReportReactionRepository,
	imageRepo reportRepository.ReportImageRepository,
	userRepo userRepository.UserRepository,
	userProfileRepo userRepository.UserProfileRepository,
	reportProgressRepo reportRepository.ReportProgressRepository,
	reportVoteRepo reportRepository.ReportVoteRepository,
	tasksService task.TaskService,
	reportCommentRepo reportRepository.ReportCommentRepository,
	reportSavedRepo reportRepository.ReportSavedRepository,
) *ReportService {
	return &ReportService{
		postgreDB:          postgreDB,
		mongoDB:            mongoDB,
		reportRepo:         reportRepo,
		reportLocationRepo: locationRepo,
		reportImageRepo:    imageRepo,
		userRepo:           userRepo,
		reportReactionRepo: reportReaction,
		reportProgressRepo: reportProgressRepo,
		userProfileRepo:    userProfileRepo,
		reportVoteRepo:     reportVoteRepo,
		tasksService:       tasksService,
		reportCommentRepo:  reportCommentRepo,
		reportSavedRepo:    reportSavedRepo,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, userID uint, req dto.CreateReportRequest) (*dto.CreateReportResponse, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("Creating report",
		zap.String("request_id", requestID),
		zap.Uint("user_id", userID),
		zap.String("report_type", req.ReportType),
	)

	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		logger.Error("Failed to start transaction",
			zap.String("request_id", requestID),
			zap.Error(tx.Error),
		)
		return nil, util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	currentUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		tx.Rollback()
		return nil, util.New(500, "USER_GET_FAILED", "Failed to retrieve user information", err.Error(), nil)
	}

	var reportStruct model.Report
	reportStruct = model.Report{
		UserID:            userID,
		ReportTitle:       req.ReportTitle,
		HasProgress:       req.HasProgress,
		ReportStatus:      model.WAITING,
		ReportType:        model.ReportType(req.ReportType),
		ReportDescription: req.ReportDescription,
		CreatedAt:         time.Now().Unix(),
		UpdatedAt:         time.Now().Unix(),
	}
	if err := s.reportRepo.Create(ctx, &reportStruct, tx); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_CREATE_FAILED", "Failed to create report", err.Error(), nil)
	}

	reportID := reportStruct.ID

	var reportLocationStruct model.ReportLocation
	reportLocationStruct = model.ReportLocation{
		ReportID:       reportID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Geometry:       fmt.Sprintf("SRID=4326;POINT(%f %f)", req.Longitude, req.Latitude),
		DetailLocation: req.DetailLocation,
		DisplayName:    req.DisplayName,
		MapZoom:        req.MapZoom,
		AddressType:    req.AddressType,
		Country:        req.Country,
		CountryCode:    req.CountryCode,
		Region:         req.Region,
		Road:           req.Road,
		PostCode:       req.PostCode,
		County:         req.County,
		State:          req.State,
		Village:        req.Village,
		Suburb:         req.Suburb,
	}

	if err := s.reportLocationRepo.Create(ctx, &reportLocationStruct, tx); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_LOCATION_CREATE_FAILED", "Failed to save report location", err.Error(), nil)
	}

	var reportImages model.ReportImage
	reportImages = model.ReportImage{
		Image1URL: req.Image1URL,
		Image2URL: req.Image2URL,
		Image3URL: req.Image3URL,
		Image4URL: req.Image4URL,
		Image5URL: req.Image5URL,
		ReportID:  reportID,
	}
	if err := s.reportImageRepo.Create(ctx, &reportImages, tx); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_IMAGE_CREATE_FAILED", "Failed to save report images", err.Error(), nil)
	}

	if err := s.tasksService.SendFollowerReportNotificationTask(reportStruct, *currentUser); err != nil {
		tx.Rollback()
		return nil, util.New(500, "AUTO_RESOLVE_TASK_FAILED", "Failed to create automatic resolution task", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}

	reportResult := &dto.CreateReportResponse{
		Report:         reportStruct,
		ReportLocation: reportLocationStruct,
		ReportImages:   reportImages,
	}

	logger.Info("Report created successfully",
		zap.String("request_id", requestID),
		zap.Uint("report_id", reportStruct.ID),
		zap.Uint("user_id", userID),
	)

	return reportResult, nil
}

func (s *ReportService) EditReport(ctx context.Context, userID, reportID uint, req dto.EditReportRequest) (*dto.EditReportResponse, error) {
	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		return nil, util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	existingReport, err := s.reportRepo.GetByIDTX(ctx, tx, reportID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	if existingReport.UserID != userID {
		tx.Rollback()
		return nil, util.New(403, "FORBIDDEN", "You do not have permission to upload progress for this report", "", nil)
	}

	if existingReport.ReportStatus == model.RESOLVED {
		tx.Rollback()
		return nil, util.New(400, "REPORT_ALREADY_RESOLVED", "The report is complete and can no longer be edited", "", nil)
	}

	existingReportLocation := existingReport.ReportLocation
	existingReportImages := existingReport.ReportImages

	switch existingReport.ReportStatus {
	case model.WAITING:
		existingReport.ReportTitle = req.ReportTitle
		existingReport.HasProgress = req.HasProgress
		existingReport.ReportType = model.ReportType(req.ReportType)
		existingReport.ReportDescription = req.ReportDescription

		existingReportLocation.Latitude = req.Latitude
		existingReportLocation.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", req.Longitude, req.Latitude)
		existingReportLocation.Longitude = req.Longitude
		existingReportLocation.DetailLocation = req.DetailLocation
		existingReportLocation.DisplayName = req.DisplayName
		existingReportLocation.AddressType = req.AddressType
		existingReportLocation.Country = req.Country
		existingReportLocation.CountryCode = req.CountryCode
		existingReportLocation.Region = req.Region
		existingReportLocation.Road = req.Road
		existingReportLocation.PostCode = req.PostCode
		existingReportLocation.County = req.County
		existingReportLocation.State = req.State
		existingReportLocation.Village = req.Village
		existingReportLocation.Suburb = req.Suburb
		existingReportLocation.MapZoom = req.MapZoom

		existingReportImages.Image1URL = req.Image1URL
		existingReportImages.Image2URL = req.Image2URL
		existingReportImages.Image3URL = req.Image3URL
		existingReportImages.Image4URL = req.Image4URL
		existingReportImages.Image5URL = req.Image5URL

	case model.ON_PROGRESS, model.WAITING_CONFIRMATION, model.EXPIRED:
		existingReport.ReportDescription = req.ReportDescription

		existingReportLocation.MapZoom = req.MapZoom
		existingReportImages.Image1URL = req.Image1URL
		existingReportImages.Image2URL = req.Image2URL
		existingReportImages.Image3URL = req.Image3URL
		existingReportImages.Image4URL = req.Image4URL
		existingReportImages.Image5URL = req.Image5URL
	}

	existingReport.UpdatedAt = time.Now().Unix()

	if _, err := s.reportRepo.UpdateTX(ctx, tx, existingReport); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_UPDATE_FAILED", "Failed to update report", err.Error(), nil)
	}

	if _, err := s.reportLocationRepo.UpdateTX(ctx, tx, existingReportLocation); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_LOCATION_UPDATE_FAILED", "Failed to update report location", err.Error(), nil)
	}

	if _, err := s.reportImageRepo.UpdateTX(ctx, tx, existingReportImages); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_IMAGE_UPDATE_FAILED", "Failed to update report images", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}

	reportResult := &dto.EditReportResponse{
		Report:         *existingReport,
		ReportLocation: *existingReportLocation,
		ReportImages:   *existingReportImages,
	}
	return reportResult, nil
}

func (s *ReportService) DeleteReport(ctx context.Context, userID, reportID uint, deleteType string) error {
	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		return util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	existingReport, err := s.reportRepo.GetByIDTX(ctx, tx, reportID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}
	if existingReport.UserID != userID {
		tx.Rollback()
		return util.New(403, "FORBIDDEN", "You do not have permission to delete this report", "", nil)
	}

	currentTime := time.Now().Unix()
	switch deleteType {
	case "soft":
		isDeleted := true
		existingReport.IsDeleted = &isDeleted
		existingReport.DeletedAt = &currentTime
		if _, err := s.reportRepo.UpdateTX(ctx, tx, existingReport); err != nil {
			tx.Rollback()
			return util.New(500, "REPORT_DELETE_FAILED", "Failed to delete report", err.Error(), nil)
		}
	case "hard":
		if _, err := s.reportRepo.DeleteTX(ctx, tx, existingReport); err != nil {
			tx.Rollback()
			return util.New(500, "REPORT_DELETE_FAILED", "Failed to delete report", err.Error(), nil)
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}
	return nil
}

func (s *ReportService) GetAllReport(ctx context.Context, userID, cursorID, reportOwnerID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance) (*dto.GetReportsResponse, error) {
	isDeleted := false
	limit := 5
	reports, err := s.reportRepo.GetByIsDeletedPaginated(ctx, uint(limit), cursorID, reportOwnerID, reportType, status, sortBy, hasProgress, distance, isDeleted)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	reportsCount, err := s.reportRepo.GetByReportTypeCount(ctx)
	if err != nil {
		return nil, util.New(500, "REPORT_COUNT_FAILED", "Failed to retrieve total report count", err.Error(), nil)
	}

	var fullReports []dto.Report

	for _, report := range *reports {
		likeReactionCount, err := s.reportReactionRepo.GetLikeReactionCount(ctx, report.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(500, "REACTION_COUNT_FAILED", "Failed to retrieve like reactions", err.Error(), nil)
		}
		dislikeReactionCount, err := s.reportReactionRepo.GetDislikeReactionCount(ctx, report.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(500, "REACTION_COUNT_FAILED", "Failed to retrieve dislike reactions", err.Error(), nil)
		}

		var isLikedByCurrentUser, isDislikedByCurrentUser, isResolvedByCurrentUser, isOnProgressByCurrentUser, isNotResolvedByCurrentUser bool

		resolvedVoteCount, err := s.reportVoteRepo.GetResolvedVoteCount(ctx, report.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve 'RESOLVED' votes", err.Error(), nil)
		}

		onProgressVoteCount, err := s.reportVoteRepo.GetOnProgressVoteCount(ctx, report.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve 'ON_PROGRESS' votes", err.Error(), nil)
		}

		reportSaved, err := s.reportSavedRepo.GetByUserIDAndReportID(ctx, userID, report.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(500, "REPORT_SAVED_FETCH_FAILED", "Failed to retrieve report save status", err.Error(), nil)
		}

		fullReports = append(fullReports, dto.Report{
			ID:                report.ID,
			ReportTitle:       report.ReportTitle,
			ReportType:        string(report.ReportType),
			ReportDescription: report.ReportDescription,
			ReportCreatedAt:   report.CreatedAt,
			UserID:            report.UserID,
			UserName:          report.User.Username,
			FullName:          report.User.FullName,
			ProfilePicture:    report.User.Profile.ProfilePicture,
			ReportSaved: &dto.ReportSaved{
				UserID:   userID,
				ReportID: report.ID,
				Save:     reportSaved != nil,
			},
			Location: dto.ReportLocation{
				DetailLocation: report.ReportLocation.DetailLocation,
				Latitude:       report.ReportLocation.Latitude,
				Longitude:      report.ReportLocation.Longitude,
				DisplayName:    report.ReportLocation.DisplayName,
				AddressType:    report.ReportLocation.AddressType,
				Country:        report.ReportLocation.Country,
				CountryCode:    report.ReportLocation.CountryCode,
				Region:         report.ReportLocation.Region,
				Road:           report.ReportLocation.Road,
				MapZoom:        report.ReportLocation.MapZoom,
				PostCode:       report.ReportLocation.PostCode,
				County:         report.ReportLocation.County,
				State:          report.ReportLocation.State,
				Village:        report.ReportLocation.Village,
				Suburb:         report.ReportLocation.Suburb,
				Geometry:       &report.ReportLocation.Geometry,
			},
			ReportStatus: string(report.ReportStatus),
			HasProgress:  report.HasProgress,
			Images: dto.ReportImage{
				Image1URL: report.ReportImages.Image1URL,
				Image2URL: report.ReportImages.Image2URL,
				Image3URL: report.ReportImages.Image3URL,
				Image4URL: report.ReportImages.Image4URL,
				Image5URL: report.ReportImages.Image5URL,
			},
			TotalLikeReactions:    &likeReactionCount,
			TotalDislikeReactions: &dislikeReactionCount,
			TotalReactions:        likeReactionCount + dislikeReactionCount,
			ReportReactions: func() []dto.ReactReportResponse {
				var reactions []dto.ReactReportResponse
				for _, reaction := range *report.ReportReactions {
					reactions = append(reactions, dto.ReactReportResponse{
						ReportID:     reaction.ReportID,
						UserID:       reaction.UserID,
						ReactionType: string(reaction.Type),
						CreatedAt:    reaction.CreatedAt,
						UpdatedAt:    reaction.UpdatedAt,
					})
					if reaction.UserID == userID {
						if reaction.Type == model.Like {
							isLikedByCurrentUser = true
						}
						if reaction.Type == model.Dislike {
							isDislikedByCurrentUser = true
						}
					}
				}
				return reactions
			}(),
			ReportProgress: func() []dto.GetProgressReportResponse {
				var progresses []dto.GetProgressReportResponse
				if report.ReportProgress != nil {
					for _, progress := range *report.ReportProgress {
						progresses = append(progresses, dto.GetProgressReportResponse{
							ReportID:    progress.ReportID,
							Status:      string(progress.Status),
							Notes:       &progress.Notes,
							Attachment1: progress.Attachment1,
							Attachment2: progress.Attachment2,
							CreatedAt:   progress.CreatedAt,
						})
					}
				}
				return progresses
			}(),
			IsLikedByCurrentUser:    isLikedByCurrentUser,
			IsDislikedByCurrentUser: isDislikedByCurrentUser,
			TotalResolvedVotes:      &resolvedVoteCount,
			TotalOnProgressVotes:    &onProgressVoteCount,
			TotalVotes:              resolvedVoteCount + onProgressVoteCount,
			ReportVotes: func() []dto.GetVoteReportResponse {
				var votes []dto.GetVoteReportResponse
				for _, vote := range *report.ReportVotes {
					votes = append(votes, dto.GetVoteReportResponse{
						ID:        vote.ID,
						ReportID:  vote.ReportID,
						UserID:    vote.UserID,
						VoteType:  vote.VoteType,
						CreatedAt: vote.CreatedAt,
						UpdatedAt: vote.UpdatedAt,
					})
					if vote.UserID == userID {
						if vote.VoteType == model.RESOLVED {
							isResolvedByCurrentUser = true
						}
						if vote.VoteType == model.ON_PROGRESS {
							isOnProgressByCurrentUser = true
						}
					}
				}
				return votes
			}(),
			IsResolvedByCurrentUser:    isResolvedByCurrentUser,
			IsNotResolvedByCurrentUser: isNotResolvedByCurrentUser,
			IsOnProgressByCurrentUser:  isOnProgressByCurrentUser,
			MajorityVote:               getMajorityVote(resolvedVoteCount, onProgressVoteCount),
			LastUpdatedBy:              (*string)(&report.LastUpdatedBy),
			LastUpdatedProgressAt:      report.LastUpdatedProgressAt,
			ReportUpdatedAt:            report.UpdatedAt,
		})
	}
	reportsData := dto.GetReportsResponse{
		Reports:     fullReports,
		TotalCounts: reportsCount,
	}
	return &reportsData, nil
}

func (s *ReportService) GetReportByID(ctx context.Context, userID, reportID uint) (*dto.GetReportResponse, error) {
	isDeleted := false
	report, err := s.reportRepo.GetByIDIsDeleted(ctx, reportID, isDeleted)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}
	var isLikedByCurrentUser, isDislikedByCurrentUser, isResolvedByCurrentUser, isOnProgressByCurrentUser bool
	likeReactionCount, err := s.reportReactionRepo.GetLikeReactionCount(ctx, report.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(500, "REACTION_COUNT_FAILED", "Failed to retrieve like reactions", err.Error(), nil)
	}
	dislikeReactionCount, err := s.reportReactionRepo.GetDislikeReactionCount(ctx, report.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(500, "REACTION_COUNT_FAILED", "Failed to retrieve dislike reactions", err.Error(), nil)
	}
	resolvedVoteCount, err := s.reportVoteRepo.GetResolvedVoteCount(ctx, report.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve 'RESOLVED' votes", err.Error(), nil)
	}
	onProgressVoteCount, err := s.reportVoteRepo.GetOnProgressVoteCount(ctx, report.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve 'ON_PROGRESS' votes", err.Error(), nil)
	}

	reportSaved, err := s.reportSavedRepo.GetByUserIDAndReportID(ctx, userID, report.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(500, "REPORT_SAVED_FETCH_FAILED", "Failed to retrieve report save status", err.Error(), nil)
	}

	fullReport := dto.Report{
		ID:                report.ID,
		ReportTitle:       report.ReportTitle,
		ReportType:        string(report.ReportType),
		ReportDescription: report.ReportDescription,
		ReportCreatedAt:   report.CreatedAt,
		UserID:            report.UserID,
		UserName:          report.User.Username,
		FullName:          report.User.FullName,
		ProfilePicture:    report.User.Profile.ProfilePicture,
		ReportSaved: &dto.ReportSaved{
			UserID:   userID,
			ReportID: report.ID,
			Save:     reportSaved != nil,
		},
		Location: dto.ReportLocation{
			DetailLocation: report.ReportLocation.DetailLocation,
			Latitude:       report.ReportLocation.Latitude,
			Longitude:      report.ReportLocation.Longitude,
			DisplayName:    report.ReportLocation.DisplayName,
			AddressType:    report.ReportLocation.AddressType,
			Country:        report.ReportLocation.Country,
			CountryCode:    report.ReportLocation.CountryCode,
			Region:         report.ReportLocation.Region,
			Road:           report.ReportLocation.Road,
			PostCode:       report.ReportLocation.PostCode,
			County:         report.ReportLocation.County,
			State:          report.ReportLocation.State,
			Village:        report.ReportLocation.Village,
			Suburb:         report.ReportLocation.Suburb,
			Geometry:       &report.ReportLocation.Geometry,
			MapZoom:        report.ReportLocation.MapZoom,
		},
		ReportStatus: string(report.ReportStatus),
		HasProgress:  report.HasProgress,
		Images: dto.ReportImage{
			Image1URL: report.ReportImages.Image1URL,
			Image2URL: report.ReportImages.Image2URL,
			Image3URL: report.ReportImages.Image3URL,
			Image4URL: report.ReportImages.Image4URL,
			Image5URL: report.ReportImages.Image5URL,
		},
		TotalLikeReactions:    &likeReactionCount,
		TotalDislikeReactions: &dislikeReactionCount,
		TotalReactions:        likeReactionCount + dislikeReactionCount,
		ReportReactions: func() []dto.ReactReportResponse {
			var reactions []dto.ReactReportResponse
			for _, reaction := range *report.ReportReactions {
				reactions = append(reactions, dto.ReactReportResponse{
					ReportID:     reaction.ReportID,
					UserID:       reaction.UserID,
					ReactionType: string(reaction.Type),
					CreatedAt:    reaction.CreatedAt,
					UpdatedAt:    reaction.UpdatedAt,
				})
				if reaction.UserID == report.UserID {
					if reaction.Type == model.Like {
						isLikedByCurrentUser = true
					}
					if reaction.Type == model.Dislike {
						isDislikedByCurrentUser = true
					}
				}
			}
			return reactions
		}(),
		ReportProgress: func() []dto.GetProgressReportResponse {
			var progresses []dto.GetProgressReportResponse
			if report.ReportProgress != nil {
				for _, progress := range *report.ReportProgress {
					progresses = append(progresses, dto.GetProgressReportResponse{
						ReportID:    progress.ReportID,
						Status:      string(progress.Status),
						Notes:       &progress.Notes,
						Attachment1: progress.Attachment1,
						Attachment2: progress.Attachment2,
						CreatedAt:   progress.CreatedAt,
					})
				}
			}
			return progresses
		}(),
		TotalResolvedVotes:   &resolvedVoteCount,
		TotalOnProgressVotes: &onProgressVoteCount,
		TotalVotes:           resolvedVoteCount + onProgressVoteCount,
		ReportVotes: func() []dto.GetVoteReportResponse {
			var votes []dto.GetVoteReportResponse
			for _, vote := range *report.ReportVotes {
				votes = append(votes, dto.GetVoteReportResponse{
					ID:        vote.ID,
					ReportID:  vote.ReportID,
					UserID:    vote.UserID,
					VoteType:  vote.VoteType,
					CreatedAt: vote.CreatedAt,
					UpdatedAt: vote.UpdatedAt,
				})
				if vote.UserID == userID {
					if vote.VoteType == model.RESOLVED {
						isResolvedByCurrentUser = true
					}
					if vote.VoteType == model.ON_PROGRESS {
						isOnProgressByCurrentUser = true
					}
				}
			}
			return votes
		}(),
		IsLikedByCurrentUser:      isLikedByCurrentUser,
		IsDislikedByCurrentUser:   isDislikedByCurrentUser,
		IsResolvedByCurrentUser:   isResolvedByCurrentUser,
		IsOnProgressByCurrentUser: isOnProgressByCurrentUser,
		MajorityVote:              getMajorityVote(resolvedVoteCount, onProgressVoteCount),
		LastUpdatedBy:             (*string)(&report.LastUpdatedBy),
		LastUpdatedProgressAt:     report.LastUpdatedProgressAt,
		ReportUpdatedAt:           report.UpdatedAt,
	}
	result := dto.GetReportResponse{
		Report: fullReport,
	}
	return &result, nil
}

func (s *ReportService) SearchReports(ctx context.Context, searchQuery string, reportsDataNextCursor uint, limit int) (*dto.SearchResponse, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("Searching reports",
		zap.String("request_id", requestID),
		zap.String("search_query", searchQuery),
		zap.Uint("next_cursor", reportsDataNextCursor),
		zap.Int("limit", limit),
	)

	reportsData, err := s.reportRepo.FullTextSearchReportPaginated(ctx, searchQuery, limit, reportsDataNextCursor)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_SEARCH_FAILED", "Failed to search report", err.Error(), nil)
	}

	resultReports := make([]dto.ReportsSearch, 0, len(*reportsData))
	for _, report := range *reportsData {
		reportDTO := dto.ReportsSearch{
			ID:                report.ID,
			ReportTitle:       report.ReportTitle,
			ReportType:        string(report.ReportType),
			ReportDescription: report.ReportDescription,
			ReportHasProgress: *report.HasProgress,
			ReportStatus:      string(report.ReportStatus),
			CreatedAt:         report.CreatedAt,
			UpdatedAt:         report.UpdatedAt,
		}
		resultReports = append(resultReports, reportDTO)
	}
	result := &dto.SearchResponse{
		ReportsData: resultReports,
	}
	return result, nil
}

func (s *ReportService) ReactToReport(ctx context.Context, userID uint, reportID uint, reactionType string) (*dto.ReactReportResponse, error) {
	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		return nil, util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	modelReactionType := model.ReactionType(reactionType)
	var resultReaction *model.ReportReaction
	isDelete := false

	existingReaction, err := s.reportReactionRepo.GetByUserReportIDTX(ctx, tx, userID, reportID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, util.New(500, "REACTION_FETCH_FAILED", "Failed to retrieve report reaction", err.Error(), nil)
	}

	switch {
	case existingReaction == nil:
		newReaction := model.ReportReaction{
			UserID:   userID,
			ReportID: reportID,
			Type:     modelReactionType,
		}
		newReportReaction, err := s.reportReactionRepo.CreateTX(ctx, tx, &newReaction)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "REACTION_CREATE_FAILED", "Failed to add reaction", err.Error(), nil)
		}
		resultReaction = newReportReaction

	case existingReaction.Type == modelReactionType:
		if err := s.reportReactionRepo.DeleteTX(ctx, tx, existingReaction); err != nil {
			tx.Rollback()
			return nil, util.New(500, "REACTION_DELETE_FAILED", "Failed to delete reaction", err.Error(), nil)
		}
		resultReaction = nil
		isDelete = true

	default:
		existingReaction.Type = modelReactionType
		existingReaction.UpdatedAt = time.Now().Unix()
		updatedReportReaction, err := s.reportReactionRepo.UpdateTX(ctx, tx, existingReaction)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "REACTION_UPDATE_FAILED", "Failed to update reaction", err.Error(), nil)
		}
		resultReaction = updatedReportReaction
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}

	response := &dto.ReactReportResponse{
		ReportID: reportID,
		UserID:   userID,
	}
	if resultReaction != nil {
		response.ReactionType = string(resultReaction.Type)
		response.CreatedAt = resultReaction.CreatedAt
		response.UpdatedAt = resultReaction.UpdatedAt
	}

	if isDelete {
		return response, nil
	}

	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	reactorUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	if report.UserID != userID {
		if err := s.tasksService.CreateNotificationTask(
			report.UserID,
			"Someone reacted to your report",
			fmt.Sprintf("User %s reacted to your report", reactorUser.Username),
			util.StrPtrOrNil(reactorUser.Username),
			model.EntityTypeUser,
			model.ReportNotificationCategory,
			model.NotificationTypeInfo,
		); err != nil {
			return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create notification task", err.Error(), nil)
		}
	}

	return response, nil
}

func (s *ReportService) VoteToReport(ctx context.Context, userID uint, reportID uint, voteType string) (*dto.GetVoteReportResponse, error) {
	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		return nil, util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	existingVote, err := s.reportVoteRepo.GetByUserReportIDTX(ctx, tx, userID, reportID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, util.New(500, "VOTE_FETCH_FAILED", "Failed to retrieve report votes", err.Error(), nil)
	}

	if existingVote != nil {
		return nil, util.New(400, "ALREADY_VOTED", "You have already voted on this report", "", nil)
	}

	if report.UserID == userID {
		tx.Rollback()
		return nil, util.New(400, "CANNOT_VOTE_OWN_REPORT", "You cannot vote on your own report", "", nil)
	}

	if report.ReportStatus == model.RESOLVED {
		tx.Rollback()
		return nil, util.New(400, "REPORT_ALREADY_RESOLVED", "You cannot vote on a completed report", "", nil)
	}

	if report.ReportStatus == model.WAITING_CONFIRMATION {
		tx.Rollback()
		return nil, util.New(400, "REPORT_WAITING_CONFIRMATION", "You cannot vote on a report that is waiting for confirmation", "", nil)
	}

	if report.ReportStatus == model.EXPIRED {
		tx.Rollback()
		return nil, util.New(400, "REPORT_EXPIRED", "You cannot vote on an expired report", "", nil)
	}

	if report.HasProgress == nil || !*report.HasProgress {
		tx.Rollback()
		return nil, util.New(400, "REPORT_NO_PROGRESS", "You cannot vote on a report without progress (information only)", "", nil)
	}

	modelVoteType := model.ReportStatus(voteType)
	var resultVote *model.ReportVote

	switch {
	case existingVote == nil:
		newVote := model.ReportVote{
			UserID:    userID,
			ReportID:  reportID,
			VoteType:  modelVoteType,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		newReportVote, err := s.reportVoteRepo.CreateTX(ctx, tx, &newVote)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "VOTE_CREATE_FAILED", "Failed to add vote", err.Error(), nil)
		}
		resultVote = newReportVote

	case existingVote.VoteType == modelVoteType:
		if err := s.reportVoteRepo.DeleteTX(ctx, tx, existingVote); err != nil {
			tx.Rollback()
			return nil, util.New(500, "VOTE_DELETE_FAILED", "Failed to delete vote", err.Error(), nil)
		}
		resultVote = nil
	default:
		existingVote.VoteType = modelVoteType
		existingVote.UpdatedAt = time.Now().Unix()
		updatedReportVote, err := s.reportVoteRepo.UpdateTX(ctx, tx, existingVote)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "VOTE_UPDATE_FAILED", "Failed to update vote", err.Error(), nil)
		}
		resultVote = updatedReportVote
	}
	if resultVote != nil {
		reportVoteCounts, err := s.reportVoteRepo.GetReportVoteCountsTX(ctx, tx, reportID)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve report vote count", err.Error(), nil)
		}

		voteTypeCountsOrder := getVoteTypeOrder(reportVoteCounts)
		totalVote, err := s.reportVoteRepo.GetTotalVoteCountTX(ctx, tx, reportID)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "VOTE_COUNT_FAILED", "Failed to retrieve total report votes", err.Error(), nil)
		}
		topVote := voteTypeCountsOrder[0]
		secondVote := voteTypeCountsOrder[1]

		marginVote := float64(topVote.Count-secondVote.Count) / float64(totalVote) * 100

		limitTopVote := 2
		if marginVote >= 20.0 && topVote.Count >= limitTopVote && report.ReportStatus != model.WAITING_CONFIRMATION && topVote.Type != report.ReportStatus {
			report.ReportStatus = model.WAITING_CONFIRMATION
			report.LastUpdatedBy = model.System
			report.LastUpdatedProgressAt = util.Int64PtrOrNil(time.Now().Unix())
			report.PotentiallyResolvedAt = util.Int64PtrOrNil(time.Now().Unix())
			reportLink := fmt.Sprintf("%s/main/reports/%d", util.ClientURL(), report.ID)
			go sendPotentiallyResolvedReportEmail(
				report.User.Email,
				report.User.Username,
				report.ReportTitle,
				reportLink,
				7,
				report.User.IsDisableEmailNotification,
			)
			if err := s.tasksService.AutoResolveReportTask(reportID); err != nil {
				tx.Rollback()
				return nil, util.New(500, "AUTO_RESOLVE_TASK_FAILED", "Failed to create automatic resolution task", err.Error(), nil)
			}

			reportStatusMessage := map[any]string{
				model.RESOLVED:             "RESOLVED",
				model.ON_PROGRESS:          "IN PROGRESS",
				model.WAITING_CONFIRMATION: "WAITING FOR CONFIRMATION",
			}
			progressNotes := fmt.Sprintf("The report is waiting for confirmation because it received the highest number of votes with report status '%s' and total votes: %d.", reportStatusMessage[topVote.Type], totalVote)

			var newProgress *model.ReportProgress
			if report.ReportStatus == model.WAITING_CONFIRMATION {
				newProgress = &model.ReportProgress{
					ReportID:  reportID,
					UserID:    userID,
					Status:    model.WAITING_CONFIRMATION,
					Notes:     progressNotes,
					CreatedAt: time.Now().Unix(),
				}
			}

			if _, err := s.reportProgressRepo.CreateTX(ctx, tx, newProgress); err != nil {
				tx.Rollback()
				return nil, util.New(500, "PROGRESS_CREATE_FAILED", "Failed to upload report progress", err.Error(), nil)
			}
		}

		if _, err := s.reportRepo.UpdateTX(ctx, tx, report); err != nil {
			tx.Rollback()
			return nil, util.New(500, "REPORT_UPDATE_FAILED", "Failed to update report status", err.Error(), nil)
		}
	}

	if resultVote != nil && report.UserID != userID {
		voterUser, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			tx.Rollback()
			return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
		}
		if err := s.tasksService.CreateNotificationTask(
			report.UserID,
			"Someone voted on your report",
			fmt.Sprintf("User %s voted on your report", voterUser.Username),
			util.StrPtrOrNil(strconv.FormatUint(uint64(reportID), 10)),
			model.EntityTypeReport,
			model.ReportNotificationCategory,
			model.NotificationTypeInfo,
		); err != nil {
			tx.Rollback()
			return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create vote notification task", err.Error(), nil)
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}

	return &dto.GetVoteReportResponse{
		ID:                    resultVote.ID,
		ReportID:              reportID,
		ReportStatus:          report.ReportStatus,
		UserID:                userID,
		VoteType:              resultVote.VoteType,
		CreatedAt:             resultVote.CreatedAt,
		UpdatedAt:             resultVote.UpdatedAt,
		LastUpdatedBy:         (*string)(&report.LastUpdatedBy),
		LastUpdatedProgressAt: report.LastUpdatedProgressAt,
	}, nil
}

func (s *ReportService) UploadProgressReport(ctx context.Context, userID, reportID uint, req dto.UploadProgressReportRequest) (*dto.UploadProgressReportResponse, error) {
	tx := s.postgreDB.Begin()
	if tx.Error != nil {
		return nil, util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		if tx.Error != nil {
			tx.Rollback()
		}
	}()

	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	if report.UserID != userID {
		tx.Rollback()
		return nil, util.New(403, "FORBIDDEN", "You do not have permission to upload progress for this report", "", nil)
	}

	if report.HasProgress == nil || !*report.HasProgress {
		tx.Rollback()
		return nil, util.New(400, "REPORT_NO_PROGRESS", "This report does not require progress (information only)", "", nil)
	}

	if report.ReportStatus == model.RESOLVED {
		tx.Rollback()
		return nil, util.New(400, "REPORT_ALREADY_RESOLVED", "The report is complete and can no longer receive progress updates", "", nil)
	}

	if req.Status == string(model.RESOLVED) {
		report.ReportStatus = model.RESOLVED
		if _, err := s.reportRepo.UpdateTX(ctx, tx, report); err != nil {
			tx.Rollback()
			return nil, util.New(500, "REPORT_UPDATE_FAILED", "Failed to update report status", err.Error(), nil)
		}
	}

	reportProgress := &model.ReportProgress{
		ReportID:    reportID,
		UserID:      userID,
		Status:      model.ReportStatus(req.Status),
		Notes:       req.Notes,
		Attachment1: req.Attachment1,
		Attachment2: req.Attachment2,
		CreatedAt:   time.Now().Unix(),
	}
	newProgress, err := s.reportProgressRepo.CreateTX(ctx, tx, reportProgress)
	if err != nil {
		tx.Rollback()
		return nil, util.New(500, "PROGRESS_CREATE_FAILED", "Failed to upload report progress", err.Error(), nil)
	}

	report.ReportStatus = model.ReportStatus(req.Status)
	report.AdminOverride = util.BoolPtrOrNil(true)
	report.LastUpdatedBy = model.Owner
	report.LastUpdatedProgressAt = util.Int64PtrOrNil(time.Now().Unix())
	if _, err := s.reportRepo.UpdateTX(ctx, tx, report); err != nil {
		tx.Rollback()
		return nil, util.New(500, "REPORT_UPDATE_FAILED", "Failed to update report status", err.Error(), nil)
	}

	response := &dto.UploadProgressReportResponse{
		ReportID:              newProgress.ReportID,
		Status:                string(newProgress.Status),
		Notes:                 &newProgress.Notes,
		Attachment1:           newProgress.Attachment1,
		Attachment2:           newProgress.Attachment2,
		CreatedAt:             newProgress.CreatedAt,
		LastUpdatedProgressAt: report.LastUpdatedProgressAt,
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save transaction", err.Error(), nil)
	}

	return response, nil
}

func (s *ReportService) GetProgressReports(ctx context.Context, reportID uint) ([]dto.GetProgressReportResponse, error) {
	reportProgresses, err := s.reportProgressRepo.GetByReportID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "PROGRESS_NOT_FOUND", "Report progress not found", "", nil)
		}
		return nil, util.New(500, "PROGRESS_FETCH_FAILED", "Failed to retrieve report progress", err.Error(), nil)
	}
	var response []dto.GetProgressReportResponse
	for _, progress := range reportProgresses {
		response = append(response, dto.GetProgressReportResponse{
			ReportID:    progress.ReportID,
			Status:      string(progress.Status),
			Notes:       &progress.Notes,
			Attachment1: progress.Attachment1,
			Attachment2: progress.Attachment2,
			CreatedAt:   progress.CreatedAt,
		})
	}
	return response, nil
}

func (s *ReportService) CreateReportComment(ctx context.Context, userID, reportID uint, req dto.CreateReportCommentRequest) (*dto.CreateReportCommentResponse, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "REPORT_NOT_FOUND", "Report not found", "", nil)
		}
		return nil, util.New(500, "REPORT_FETCH_FAILED", "Failed to retrieve report", err.Error(), nil)
	}

	if report.IsDeleted != nil && *report.IsDeleted {
		return nil, util.New(400, "REPORT_DELETED", "Cannot comment on a deleted report", "", nil)
	}

	if report.ReportStatus == model.EXPIRED {
		return nil, util.New(400, "REPORT_EXPIRED", "Cannot comment on an expired report", "", nil)
	}

	parentCommentIDObj, err := util.StringPtrToObjectIDPtr(req.ParentCommentID)
	if err != nil {
		return nil, util.New(400, "INVALID_PARENT_COMMENT_ID", "Invalid parent comment ID", err.Error(), nil)
	}

	threadRootIDObj, err := util.StringPtrToObjectIDPtr(req.ThreadRootID)
	if err != nil {
		return nil, util.New(400, "INVALID_THREAD_ROOT_ID", "Invalid thread root ID", err.Error(), nil)
	}

	var commentMedia model.CommentMedia

	if req.MediaType != nil {
		commentMediaType := model.CommentMediaType(*req.MediaType)
		if req.MediaURL == nil {
			return nil, util.New(400, "MEDIA_URL_REQUIRED", "Media URL is required when a media type is provided", "", nil)
		}
		commentMedia = model.CommentMedia{
			URL:    *req.MediaURL,
			Type:   commentMediaType,
			Width:  req.MediaWidth,
			Height: req.MediaHeight,
		}
	}

	var reportComment model.ReportComment
	reportComment = model.ReportComment{
		ReportID:        reportID,
		UserID:          userID,
		Content:         req.Content,
		CreatedAt:       time.Now().Unix(),
		Media:           &commentMedia,
		UpdatedAt:       util.Int64PtrOrNil(time.Now().Unix()),
		ParentCommentID: parentCommentIDObj,
		ThreadRootID:    threadRootIDObj,
	}

	reportCommentCreated, err := s.reportCommentRepo.Create(ctx, &reportComment)
	if err != nil {
		return nil, util.New(500, "COMMENT_CREATE_FAILED", "Failed to create report comment", err.Error(), nil)
	}
	newCommentID := reportCommentCreated.ID.Hex()

	commenter, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	if report.UserID != userID {
		entityID := fmt.Sprintf("%d:%s", reportID, newCommentID)
		if err := s.tasksService.CreateNotificationTask(
			report.UserID,
			"Someone commented on your report",
			fmt.Sprintf("User %s commented on your report", commenter.Username),
			util.StrPtrOrNil(entityID),
			model.EntityTypeComment,
			model.ReportNotificationCategory,
			model.NotificationTypeInfo,
		); err != nil {
			return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create report notification task", err.Error(), nil)
		}

		if err := s.tasksService.SendReportCommentNotificationTask(*report, *reportCommentCreated, *commenter); err != nil {
			return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create report email notification task", err.Error(), nil)
		}
	}

	if parentCommentIDObj != nil {
		parentComment, err := s.reportCommentRepo.GetByID(ctx, *parentCommentIDObj)
		if err == nil && parentComment.UserID != userID && parentComment.UserID != report.UserID {
			entityID := fmt.Sprintf("%d:%s", reportID, newCommentID)
			if err := s.tasksService.CreateNotificationTask(
				parentComment.UserID,
				"Seseorang membalas komentar Anda",
				fmt.Sprintf("User %s replied to your comment", commenter.Username),
				util.StrPtrOrNil(entityID),
				model.EntityTypeComment,
				model.ReportNotificationCategory,
				model.NotificationTypeInfo,
			); err != nil {
				return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create reply notification task", err.Error(), nil)
			}
		}
	}

	for _, mentionedUserID := range req.Mentions {
		if mentionedUserID == userID || mentionedUserID == report.UserID {
			continue
		}
		entityID := fmt.Sprintf("%d:%s", reportID, newCommentID)
		if err := s.tasksService.CreateNotificationTask(
			mentionedUserID,
			fmt.Sprintf("User %s mentioned you in a comment", commenter.Username),
			fmt.Sprintf("/reports/%d", reportID),
			util.StrPtrOrNil(entityID),
			model.EntityTypeComment,
			model.ReportNotificationCategory,
			model.NotificationTypeInfo,
		); err != nil {
			return nil, util.New(500, "NOTIFICATION_TASK_FAILED", "Failed to create mention notification task", err.Error(), nil)
		}
	}

	var parentCommentIDStr, threadRootIDStr *string
	if reportCommentCreated.ParentCommentID != nil {
		hexValue := reportCommentCreated.ParentCommentID.Hex()
		parentCommentIDStr = &hexValue
	}
	if reportCommentCreated.ThreadRootID != nil {
		hexValue := reportCommentCreated.ThreadRootID.Hex()
		threadRootIDStr = &hexValue
	}
	return &dto.CreateReportCommentResponse{
		CommentID:       newCommentID,
		ReportID:        reportCommentCreated.ReportID,
		UserID:          reportCommentCreated.UserID,
		CreatedAt:       reportComment.CreatedAt,
		Content:         reportComment.Content,
		ParentCommentID: parentCommentIDStr,
		ThreadRootID:    threadRootIDStr,
	}, nil
}

func (s *ReportService) GetReportComments(ctx context.Context, reportID uint, cursorID *string) (*dto.GetReportCommentsResponse, error) {
	limit := 50

	primitiveCursor, err := util.StringPtrToObjectIDPtr(cursorID)
	if err != nil {
		return nil, util.New(400, "INVALID_CURSOR_ID", "Invalid cursor ID", err.Error(), nil)
	}

	commentsFromDB, err := s.reportCommentRepo.GetPaginatedRootByReportID(ctx, reportID, primitiveCursor, (limit + 1))
	if err != nil {
		return nil, util.New(500, "COMMENT_FETCH_FAILED", "Failed to retrieve comments", err.Error(), nil)
	}

	hasMore := len(commentsFromDB) > limit
	if hasMore {
		commentsFromDB = commentsFromDB[:limit]
	}

	userIDs := make([]uint, 0)
	userIDMap := make(map[uint]struct{})

	for _, c := range commentsFromDB {
		if _, exists := userIDMap[c.UserID]; !exists {
			userIDs = append(userIDs, c.UserID)
			userIDMap[c.UserID] = struct{}{}
		}
	}

	regexMentionPattern := regexp.MustCompile(`\[mention:(\d+)\]`)

	for _, c := range commentsFromDB {
		matchedMentions := regexMentionPattern.FindAllStringSubmatch(*c.Content, -1)
		for _, match := range matchedMentions {
			if len(match) < 2 {
				continue
			}
			mentionedID, err := strconv.ParseUint(match[1], 10, 64)
			if err != nil {
				continue
			}
			id := uint(mentionedID)
			if _, exists := userIDMap[id]; !exists {
				userIDs = append(userIDs, id)
				userIDMap[id] = struct{}{}
			}
		}
	}
	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	userMap := make(map[uint]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	replyCounts := make(map[string]int64)
	mentionsMap := make(map[string][]model.Mention)

	for _, comment := range commentsFromDB {
		count, err := s.reportCommentRepo.GetCountsByRootID(ctx, comment.ID)
		if err == nil {
			replyCounts[comment.ID.Hex()] = count
		}

		matchedMentions := regexMentionPattern.FindAllStringSubmatch(*comment.Content, -1)
		seen := make(map[uint]struct{})

		for _, match := range matchedMentions {
			if len(match) < 2 {
				continue
			}

			mentionedID, err := strconv.ParseUint(match[1], 10, 64)
			if err != nil {
				continue
			}

			id := uint(mentionedID)
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}

			if u, ok := userMap[id]; ok {
				mentionsMap[comment.ID.Hex()] = append(mentionsMap[comment.ID.Hex()], model.Mention{
					UserID:   u.ID,
					Username: u.Username,
				})
			}
		}
	}
	comments := convertRootCommentsToDTO(commentsFromDB, userMap, replyCounts, mentionsMap)

	resp := dto.GetReportCommentsResponse{
		Comments: comments,
	}

	total, _ := s.reportCommentRepo.GetCountsByReportID(ctx, reportID)
	resp.TotalCounts = total
	resp.HasMore = hasMore

	return &resp, nil
}

func (s *ReportService) GetReportStatistics(ctx context.Context) (*dto.GetReportStatisticsResponse, error) {
	totalReports, err := s.reportRepo.GetByReportTypeCount(ctx)
	if err != nil {
		return nil, util.New(500, "TOTAL_REPORTS_FETCH_FAILED", "Failed to retrieve total reports", err.Error(), nil)
	}

	reportsByStatus, err := s.reportRepo.GetByReportStatusCount(ctx, string(model.WAITING), string(model.ON_PROGRESS), string(model.WAITING_CONFIRMATION), string(model.RESOLVED), string(model.EXPIRED))
	if err != nil {
		return nil, util.New(500, "RESOLVED_REPORTS_FETCH_FAILED", "Failed to retrieve resolved reports", err.Error(), nil)
	}

	monthlyReports, err := s.reportRepo.GetMonthlyReportCount(ctx)
	if err != nil {
		return nil, util.New(500, "MONTHLY_REPORTS_FETCH_FAILED", "Failed to retrieve monthly reports", err.Error(), nil)
	}

	return &dto.GetReportStatisticsResponse{
		TotalReports:        totalReports.TotalReports,
		ReportsByStatus:     reportsByStatus,
		MonthlyReportCounts: monthlyReports,
	}, nil
}

func (s *ReportService) GetReportCommentReplies(ctx context.Context, rootID string, cursorID *string) (*dto.GetReportCommentRepliesResponse, error) {
	const limit = 60

	primitiveRootID, err := util.StringPtrToObjectIDPtr(&rootID)
	if err != nil {
		return nil, util.New(400, "INVALID_ROOT_ID", "Invalid thread root ID", err.Error(), nil)
	}

	primitiveCursor, err := util.StringPtrToObjectIDPtr(cursorID)
	if err != nil {
		return nil, util.New(400, "INVALID_CURSOR_ID", "Invalid cursor ID", err.Error(), nil)
	}

	repliesFromDB, err := s.reportCommentRepo.GetPaginatedRepliesByRootID(ctx, *primitiveRootID, primitiveCursor, limit+1)
	if err != nil {
		return nil, util.New(500, "REPLY_FETCH_FAILED", "Failed to retrieve replies", err.Error(), nil)
	}

	rootComment, err := s.reportCommentRepo.GetByID(ctx, *primitiveRootID)
	if err != nil {
		return nil, util.New(500, "ROOT_COMMENT_FETCH_FAILED", "Failed to retrieve root comment", err.Error(), nil)
	}

	userIDSet := make(map[uint]struct{})
	parentIDSet := make(map[primitive.ObjectID]struct{})
	mentionsMap := make(map[string][]uint)

	userIDSet[rootComment.UserID] = struct{}{}

	regexMentionPattern := regexp.MustCompile(`\[mention:(\d+)\]`)

	for _, reply := range repliesFromDB {
		matchedMentions := regexMentionPattern.FindAllStringSubmatch(*reply.Content, -1)
		for _, match := range matchedMentions {
			if len(match) < 2 {
				continue
			}
			userID, err := strconv.ParseUint(match[1], 10, 32)
			if err != nil {
				continue
			}
			userIDSet[uint(userID)] = struct{}{}
			mentionsMap[reply.ID.Hex()] = append(mentionsMap[reply.ID.Hex()], uint(userID))
		}
		userIDSet[reply.UserID] = struct{}{}

		if reply.ParentCommentID != nil {
			parentIDSet[*reply.ParentCommentID] = struct{}{}
		}
	}

	parentIDs := make([]primitive.ObjectID, 0, len(parentIDSet))
	for id := range parentIDSet {
		parentIDs = append(parentIDs, id)
	}

	parentsData := make(map[string]*model.ReportComment)
	if len(parentIDs) > 0 {
		parents, err := s.reportCommentRepo.GetByIDs(ctx, parentIDs)
		if err != nil {
			return nil, util.New(500, "PARENT_FETCH_FAILED", "Failed to retrieve parent comments", err.Error(), nil)
		}

		for i := range parents {
			parentsData[parents[i].ID.Hex()] = parents[i]
			userIDSet[parents[i].UserID] = struct{}{}
		}
	}

	userIDs := make([]uint, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	userMap := make(map[uint]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	hasMore := len(repliesFromDB) > limit
	if hasMore {
		repliesFromDB = repliesFromDB[:limit]
	}

	replies := convertRepliesToDTO(repliesFromDB, userMap, parentsData, mentionsMap)

	total, err := s.reportCommentRepo.GetCountsByRootID(ctx, *primitiveRootID)
	if err != nil {
		return nil, util.New(500, "COUNT_FETCH_FAILED", "Failed to count total replies", err.Error(), nil)
	}

	return &dto.GetReportCommentRepliesResponse{
		Replies:     replies,
		HasMore:     hasMore,
		TotalCounts: total,
	}, nil
}

func (s *ReportService) SaveReport(ctx context.Context, userID uint, reportID uint, save bool) (*dto.SaveReportResponse, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(
				404,
				"USER_NOT_FOUND",
				"User not found",
				"",
				nil,
			)
		}

		return nil, util.New(
			500,
			"USER_FETCH_FAILED",
			"Failed to retrieve user data",
			err.Error(),
			nil,
		)
	}

	_, err = s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(
				404,
				"REPORT_NOT_FOUND",
				"Report not found",
				"",
				nil,
			)
		}

		return nil, util.New(
			500,
			"REPORT_FETCH_FAILED",
			"Failed to retrieve report",
			err.Error(),
			nil,
		)
	}

	savedReport, err := s.reportSavedRepo.GetByUserIDAndReportID(
		ctx,
		userID,
		reportID,
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.New(
			500,
			"SAVED_REPORT_FETCH_FAILED",
			"Failed to retrieve saved report data",
			err.Error(),
			nil,
		)
	}

	tx := s.postgreDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, util.New(
			500,
			"TRANSACTION_START_FAILED",
			"Failed to start transaction",
			tx.Error.Error(),
			nil,
		)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if save {
		if savedReport == nil {
			err := s.reportSavedRepo.CreateTX(
				ctx,
				tx,
				&model.ReportSaved{
					UserID:   userID,
					ReportID: reportID,
				},
			)

			if err != nil {
				tx.Rollback()

				return nil, util.New(
					500,
					"SAVED_REPORT_CREATE_FAILED",
					"Failed to save report",
					err.Error(),
					nil,
				)
			}
		}
	} else {
		if savedReport != nil {
			err := s.reportSavedRepo.DeleteTX(
				ctx,
				tx,
				savedReport,
			)

			if err != nil {
				tx.Rollback()

				return nil, util.New(
					500,
					"SAVED_REPORT_DELETE_FAILED",
					"Failed to delete saved report",
					err.Error(),
					nil,
				)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, util.New(
			500,
			"TRANSACTION_COMMIT_FAILED",
			"Failed to save report changes",
			err.Error(),
			nil,
		)
	}

	return &dto.SaveReportResponse{
		ReportID: reportID,
		UserID:   userID,
		Save:     save,
	}, nil
}

func (s *ReportService) GetSavedReports(ctx context.Context, userID uint, cursorID *string) (*dto.GetSavedReportsResponse, error) {
	isDeleted := false
	limit := 20

	savedReports, err := s.reportSavedRepo.GetByUserIDIsDeletedPaginated(ctx, userID, isDeleted, limit, cursorID)
	if err != nil {
		return nil, util.New(
			500,
			"SAVED_REPORTS_FETCH_FAILED",
			"Failed to retrieve saved reports",
			err.Error(),
			nil,
		)
	}

	return &dto.GetSavedReportsResponse{
		SavedReports: savedReports,
	}, nil
}
