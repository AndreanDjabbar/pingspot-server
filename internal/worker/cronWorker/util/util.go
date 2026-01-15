package util

import (
	"pingspot/internal/domain/model"
	mainutils "pingspot/pkg/utils/mainUtils"
	"time"
)

func GetAutoResolvedRemainingDay(report model.Report) int {
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

func SendAutoResolvedRemainingDayEmail(to, username, reportTitle, reportLink string, daysRemaining int) error {
	return mainutils.SendEmail(mainutils.EmailData{
		To:            to,
		Subject:       "Pemberitahuan: Laporan Anda Akan Otomatis Ditandai Sebagai Terselesaikan",
		RecipientName: username,
		EmailType:     mainutils.EmailTypeProgressReminder,
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
