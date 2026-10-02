package util

import (
	reportDTO "pingspot/internal/domain/report_service/dto"
	"pingspot/internal/domain/user_service/dto"
	"pingspot/internal/model"
	mainutils "pingspot/pkg/utils/main_util"
	"sort"
)

func GetMajorityVote(resolvedVote, onProgressVote int64) *string {
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
	}

	sort.Slice(votes, func(i, j int) bool {
		return votes[i].Count > votes[j].Count
	})

	return votes
}

func SendPotentiallyResolvedReportEmail(to, username, reportTitle, reportLink string, daysRemaining int, disabledNotifications bool) error {
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
		DisabledNotifications: disabledNotifications,
		BodyTempate: getProgressReminderEmailTemplate(),
	})
}

func getProgressReminderEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="id" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Pengingat Progress Laporan</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			Laporan "{{.ReportTitle}}" perlu diperbarui. Unggah bukti progress sebelum periode berakhir.
		</div>
	
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f4f3fb;">
			<tr>
				<td align="center" style="padding: 40px 16px;">
	
					<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; width: 100%; background-color: #ffffff; border-radius: 20px; box-shadow: 0 12px 32px rgba(108, 92, 231, 0.12); overflow: hidden;">
	
						<!-- Header -->
						<tr>
							<td align="center" bgcolor="#6C5CE7" style="background-color: #6C5CE7; background-image: linear-gradient(135deg, #7B6CF0 0%, #6C5CE7 50%, #5B4BD5 100%); padding: 44px 40px 36px;">
								<h1 style="margin: 0; color: #ffffff; font-size: 30px; font-weight: 800; letter-spacing: -0.5px;">
									PingSpot
								</h1>
								<p style="margin: 10px 0 0; color: #e4e0ff; font-size: 15px; font-weight: 500; letter-spacing: 0.3px;">
									Pengingat Progress Laporan
								</p>
							</td>
						</tr>
	
						<!-- Body -->
						<tr>
							<td style="padding: 44px 40px 20px;">
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Halo {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 28px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									Laporan Anda masih berstatus
									<span style="display: inline-block; padding: 2px 10px; background-color: #fff4d6; color: #92400e; border-radius: 999px; font-size: 14px; font-weight: 600;">Dalam Peninjauan</span>
									dan menunggu pembaruan dari Anda.
								</p>
	
								<!-- Report card -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-radius: 14px;">
									<tr>
										<td style="padding: 22px 24px;">
											<p style="margin: 0 0 6px; color: #6C5CE7; font-size: 12px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase;">
												Laporan Anda
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
												⏰ Sisa waktu: 7 minggu
											</p>
											<p style="margin: 0; color: #78350f; font-size: 14px; line-height: 1.6;">
												Unggah bukti progress sebelum periode ini berakhir. Jika tidak ada pembaruan, laporan akan
												<strong>otomatis ditandai sebagai Terselesaikan</strong>.
											</p>
										</td>
									</tr>
								</table>
	
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 36px auto 14px;">
									<tr>
										<td align="center" bgcolor="#6C5CE7" style="background-color: #6C5CE7; border-radius: 12px; box-shadow: 0 6px 18px rgba(108, 92, 231, 0.35);">
											<a href="{{.ReportLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												Perbarui Progress Sekarang &rarr;
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
												Jika Anda memiliki pertanyaan, jangan ragu untuk menghubungi kami.
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
									© 2026 PingSpot. Hak cipta dilindungi undang-undang.
								</p>
								<p style="margin: 0; color: #9498b3; font-size: 12px; line-height: 1.6;">
									Ada pertanyaan? Hubungi kami melalui email
									<a href="mailto:andreanjabar18@gmail.com" style="color: #6C5CE7; text-decoration: none; font-weight: 600;">andreanjabar18@gmail.com</a>
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
		Mentions: c.Mentions,
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
		comment.Media = &model.CommentMedia{
			URL:    c.Media.URL,
			Type:   model.CommentMediaType(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return comment
}

func convertToReplyDTO(c *model.ReportComment, u *model.User, mentions []model.Mention) *reportDTO.CommentReply {
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
		Mentions: mentions,
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
		reply.Media = &model.CommentMedia{
			URL:    c.Media.URL,
			Type:   model.CommentMediaType(c.Media.Type),
			Width:  c.Media.Width,
			Height: c.Media.Height,
		}
	}
	return reply
}

func ConvertRootCommentsToDTO(comments []*model.ReportComment, users map[uint]*model.User, replyCounts map[string]int64, mentionsMap map[string][]model.Mention) []*reportDTO.Comment {
	rootComments := make([]*reportDTO.Comment, 0)

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

func ConvertRepliesToDTO(comments []*model.ReportComment, users map[uint]*model.User, parentsComment map[string]*model.ReportComment, commentMentionsMap map[string][]uint) []*reportDTO.CommentReply {
	replies := make([]*reportDTO.CommentReply, 0, len(comments))

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
