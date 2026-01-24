package util

import (
	reportDTO "pingspot/internal/domain/reportService/dto"
	"pingspot/internal/domain/userService/dto"
	"pingspot/internal/model"
	mainutils "pingspot/pkg/utils/mainUtils"
	"sort"
)

func GetMajorityVote(resolvedVote, onProgressVote, notResolvedVote int64) *string {
	if resolvedVote >= onProgressVote && resolvedVote >= notResolvedVote {
		vote := "RESOLVED"
		return &vote
	}
	if onProgressVote >= resolvedVote && onProgressVote >= notResolvedVote {
		vote := "ON_PROGRESS"
		return &vote
	}
	if notResolvedVote >= resolvedVote && notResolvedVote >= onProgressVote {
		vote := "NOT_RESOLVED"
		return &vote
	}
	return nil
}

func GetVoteTypeOrder(voteCount map[model.ReportStatus]int64) []struct {
	Type  model.ReportStatus
	Count int
} {
	votes := []struct {
		Type  model.ReportStatus
		Count int
	}{
		{model.RESOLVED, int(voteCount[model.RESOLVED])},
		{model.ON_PROGRESS, int(voteCount[model.ON_PROGRESS])},
		{model.NOT_RESOLVED, int(voteCount[model.NOT_RESOLVED])},
	}

	sort.Slice(votes, func(i, j int) bool {
		return votes[i].Count > votes[j].Count
	})

	return votes
}

func SendPotentiallyResolvedReportEmail(to, username, reportTitle, reportLink string, daysRemaining int) error {
	return mainutils.SendEmail(mainutils.EmailData{
		To:            to,
		Subject:       "Pengingat: Perbarui Progress Laporan Anda",
		RecipientName: username,
		EmailType:     mainutils.EmailTypeProgressReminder,
		TemplateData: map[string]any{
			"ReportTitle":   reportTitle,
			"ReportLink":    reportLink,
			"DaysRemaining": daysRemaining,
		},
		BodyTempate: getProgressReminderEmailTemplate(),
	})
}

func getProgressReminderEmailTemplate() string {
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
								Laporan Anda berstatus <strong style="color: #f59e0b;">Dalam Peninjauan</strong> dan perlu diperbarui!
							</p>
							<div style="margin: 30px 0; padding: 25px; background-color: #fef3c7; border-radius: 12px; border-left: 4px solid #f59e0b;">
								<p style="margin: 0 0 15px; color: #92400e; font-size: 16px; font-weight: 600;">
									📋 {{.ReportTitle}}
								</p>
								<p style="margin: 0; color: #78350f; font-size: 14px; line-height: 1.6;">
									⏰ Anda memiliki <strong>7 minggu tersisa</strong> untuk mengunggah bukti progress.<br>
									Jika tidak ada pembaruan, laporan akan <strong>otomatis ditandai sebagai Terselesaikan</strong> setelah periode ini berakhir.
								</p>
							</div>
							<div style="text-align: center; margin: 35px 0;">
								<a href="{{.ReportLink}}" 
								   style="display: inline-block; background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: #ffffff; text-decoration: none; padding: 16px 32px; border-radius: 50px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 15px rgba(245, 158, 11, 0.4); transition: all 0.3s ease; text-align: center; min-width: 200px;">
									Perbarui Progress Sekarang
								</a>
							</div>
							<div style="margin: 30px 0; text-align: center;">
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.6;">
									Jika Anda memiliki pertanyaan, jangan ragu untuk menghubungi kami.
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
								<a href="mailto:support@pingspot.com" style="color: #667eea; text-decoration: none;">
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

func convertToDTO(c *model.ReportComment, u *model.User) *reportDTO.Comment {
	comment := &reportDTO.Comment{
		CommentID: c.ID.Hex(),
		ReportID:  c.ReportID,
		UserInformation: dto.UserProfile{
			UserID:         c.UserID,
			Username:       u.Username,
			FullName:       u.FullName,
			ProfilePicture: u.Profile.ProfilePicture,
			Gender:         u.Profile.Gender,
			Bio:            u.Profile.Bio,
			Birthday:       u.Profile.Birthday,
		},
		Content: c.Content,
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
		comment.Media = &reportDTO.CommentMedia{
			URL:    c.Media.URL,
			Type:   string(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return comment
}

func convertToReplyDTO(c *model.ReportComment, u *model.User) *reportDTO.CommentReply {
	reply := &reportDTO.CommentReply{
		CommentID: c.ID.Hex(),
		ReportID:  c.ReportID,
		UserInformation: dto.UserProfile{
			UserID:         c.UserID,
			Username:       u.Username,
			FullName:       u.FullName,
			ProfilePicture: u.Profile.ProfilePicture,
			Gender:         u.Profile.Gender,
			Bio:            u.Profile.Bio,
			Birthday:       u.Profile.Birthday,
		},
		Content: c.Content,
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
		reply.Media = &reportDTO.CommentMedia{
			URL:    c.Media.URL,
			Type:   string(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return reply
}

func ConvertRootCommentsToDTO(comments []*model.ReportComment, users map[uint]*model.User, replyCounts map[string]int64) []*reportDTO.Comment {
	rootComments := make([]*reportDTO.Comment, 0)

	for _, comment := range comments {
		user := users[comment.UserID]
		if user == nil {
			continue
		}
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

func convertUserToProfile(u *model.User) *dto.UserProfile {
	return &dto.UserProfile{
		UserID:         u.ID,
		Username:       u.Username,
		FullName:       u.FullName,
		ProfilePicture: u.Profile.ProfilePicture,
		Gender:         u.Profile.Gender,
		Bio:            u.Profile.Bio,
		Birthday:       u.Profile.Birthday,
	}
}

func ConvertRepliesToDTO(comments []*model.ReportComment, users map[uint]*model.User, parentsComment map[string]*model.ReportComment) []*reportDTO.CommentReply {
	replies := make([]*reportDTO.CommentReply, 0, len(comments))

	for _, comment := range comments {
		user := users[comment.UserID]
		if user == nil {
			continue
		}

		replyDTO := convertToReplyDTO(comment, user)

		if comment.ParentCommentID != nil {
			parentID := comment.ParentCommentID.Hex()
			if parentComment, exists := parentsComment[parentID]; exists {
				if parentUser, userExists := users[parentComment.UserID]; userExists {
					if replyDTO.UserInformation.UserID != parentUser.ID {
						replyDTO.ReplyTo = convertUserToProfile(parentUser)
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
