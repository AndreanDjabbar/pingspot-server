package util

import (
	"fmt"
	"pingspot/internal/model"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/main_util"

	"go.uber.org/zap"
)

func SendNotificationNewReportEmails(report model.Report, creator model.User, followers []model.User, reportLink string) error {
	for _, follower := range followers {
		emailData := mainutils.EmailData{
			To:            follower.Email,
			Subject:       fmt.Sprintf("%s Publish a new report!", creator.Username),
			RecipientName: follower.Username,
			EmailType:     mainutils.EmailTypeNewReport,
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

		if err := mainutils.SendEmail(emailData); err != nil {
			logger.Error("Failed to send email", zap.String("email", follower.Email), zap.Error(err))
		}
	}
	return nil
}

func getNewReportEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="id" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Laporan Baru di PingSpot</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			{{.UserName}} baru saja mempublikasikan laporan: {{.ReportTitle}}
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
									Laporan Baru dari Orang yang Anda Ikuti
								</p>
							</td>
						</tr>
	
						<!-- Body -->
						<tr>
							<td style="padding: 44px 40px 20px;">
	
								<!-- Icon badge -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 22px;">
									<tr>
										<td align="center" width="64" height="64" bgcolor="#F1EFFD" style="width: 64px; height: 64px; background-color: #f1effd; border: 1px solid #ddd9fa; border-radius: 50%; font-size: 28px; line-height: 64px;">
											📢
										</td>
									</tr>
								</table>
	
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Ada laporan baru! 👋
								</h2>
								<p style="margin: 0 0 28px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									<strong style="color: #6C5CE7;">{{.UserName}}</strong> baru saja mempublikasikan laporan baru. Lihat detailnya sekarang.
								</p>
	
								<!-- Report card -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-radius: 14px;">
									<tr>
										<td style="padding: 22px 24px;">
											<p style="margin: 0 0 6px; color: #6C5CE7; font-size: 12px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase;">
												Laporan Baru
											</p>
											<p style="margin: 0 0 10px; color: #1e1b3a; font-size: 18px; font-weight: 700; line-height: 1.4;">
												📋 {{.ReportTitle}}
											</p>
											<p style="margin: 0; color: #64688a; font-size: 14px;">
												Dipublikasikan oleh <strong style="color: #3b3866;">{{.UserName}}</strong>
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
												Lihat Laporan &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<p style="margin: 0 0 8px; color: #8a8fa8; font-size: 12px; text-align: center; line-height: 1.6;">
									Tombol tidak berfungsi? Salin tautan ini ke browser Anda:<br>
									<a href="{{.ReportLink}}" style="color: #6C5CE7; text-decoration: underline; word-break: break-all;">{{.ReportLink}}</a>
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
												🔔 Anda menerima email ini karena mengikuti <strong>{{.UserName}}</strong> di PingSpot.<br>
												Anda dapat menonaktifkan notifikasi email kapan saja melalui pengaturan akun.
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
									Pertanyaan? Hubungi kami di
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