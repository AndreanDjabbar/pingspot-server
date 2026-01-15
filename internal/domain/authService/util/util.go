package util

import mainutils "pingspot/pkg/utils/mainUtils"

func SendVerificationEmail(to, username, verificationLink string) error {
	return mainutils.SendEmail(mainutils.EmailData{
		To:            to,
		Subject:       "Verifikasi Akun PingSpot",
		RecipientName: username,
		BodyTempate: getVerificationEmailTemplate(),
		EmailType:     mainutils.EmailTypeVerification,
		TemplateData: map[string]interface{}{
			"VerificationLink": verificationLink,
		},
	})
}

func getVerificationEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Verifikasi Akun PingSpot</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif; background-color: #f8fafc; line-height: 1.6;">
	<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f8fafc;">
		<tr>
			<td align="center" style="padding: 40px 20px;">
				<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1); overflow: hidden;">
					<tr>
						<td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 40px 30px; text-align: center;">
							<h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 700; letter-spacing: -0.5px;">
								PingSpot
							</h1>
							<p style="margin: 8px 0 0; color: rgba(255, 255, 255, 0.9); font-size: 16px; font-weight: 400;">
								Selamat datang di dunia terhubung Anda
							</p>
						</td>
					</tr>
					<tr>
						<td style="padding: 50px 40px;">
							<h2 style="margin: 0 0 20px; color: #1e293b; font-size: 24px; font-weight: 600; text-align: center;">
								Halo {{.UserName}}! 👋
							</h2>
							<p style="margin: 0 0 25px; color: #475569; font-size: 16px; text-align: center; line-height: 1.7;">
								Terima kasih telah bergabung dengan PingSpot! Untuk mulai menggunakan dan mengamankan akun Anda, silakan verifikasi alamat email Anda.
							</p>
							<div style="text-align: center; margin: 35px 0;">
								<a href="{{.VerificationLink}}" 
								   style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; padding: 16px 32px; border-radius: 50px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4); transition: all 0.3s ease; text-align: center; min-width: 200px;">
									Verifikasi Akun Saya
								</a>
							</div>
							<div style="margin: 30px 0; padding: 20px; background-color: #f1f5f9; border-radius: 12px; border-left: 4px solid #667eea;">
								<p style="margin: 0 0 10px; color: #475569; font-size: 14px; font-weight: 600;">
									Tombol tidak berfungsi?
								</p>
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.5;">
									Salin dan tempel link ini ke browser Anda:
								</p>
								<p style="margin: 8px 0 0; word-break: break-all;">
									<a href="{{.VerificationLink}}" style="color: #667eea; text-decoration: none; font-size: 14px;">
										{{.VerificationLink}}
									</a>
								</p>
							</div>
							<div style="margin: 30px 0; text-align: center;">
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.6;">
									🔒 Link ini akan kedaluwarsa dalam 5 menit demi keamanan Anda.<br>
									Jika Anda tidak membuat akun, abaikan email ini.
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


func SendPasswordResetEmail(to, username, resetLink string) error {
	return mainutils.SendEmail(mainutils.EmailData{
		To:            to,
		Subject:       "Reset Password PingSpot",
		RecipientName: username,
		BodyTempate: getPasswordResetEmailTemplate(),
		EmailType:     mainutils.EmailTypePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetLink": resetLink,
		},
	})
}

func getPasswordResetEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Reset Password PingSpot</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif; background-color: #f8fafc; line-height: 1.6;">
	<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f8fafc;">
		<tr>
			<td align="center" style="padding: 40px 20px;">
				<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1); overflow: hidden;">
					<tr>
						<td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 40px 30px; text-align: center;">
							<h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 700; letter-spacing: -0.5px;">
								PingSpot
							</h1>
							<p style="margin: 8px 0 0; color: rgba(255, 255, 255, 0.9); font-size: 16px; font-weight: 400;">
								Reset Password Anda
							</p>
						</td>
					</tr>
					<tr>
						<td style="padding: 50px 40px;">
							<h2 style="margin: 0 0 20px; color: #1e293b; font-size: 24px; font-weight: 600; text-align: center;">
								Halo {{.UserName}}! 👋
							</h2>
							<p style="margin: 0 0 25px; color: #475569; font-size: 16px; text-align: center; line-height: 1.7;">
								Kami menerima permintaan untuk mereset password akun Anda. Klik tombol di bawah untuk melanjutkan proses reset password.
							</p>
							<div style="text-align: center; margin: 35px 0;">
								<a href="{{.ResetLink}}" 
								   style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; padding: 16px 32px; border-radius: 50px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4); transition: all 0.3s ease; text-align: center; min-width: 200px;">
									Reset Password
								</a>
							</div>
							<div style="margin: 30px 0; padding: 20px; background-color: #f1f5f9; border-radius: 12px; border-left: 4px solid #667eea;">
								<p style="margin: 0 0 10px; color: #475569; font-size: 14px; font-weight: 600;">
									Tombol tidak berfungsi?
								</p>
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.5;">
									Salin dan tempel link ini ke browser Anda:
								</p>
								<p style="margin: 8px 0 0; word-break: break-all;">
									<a href="{{.ResetLink}}" style="color: #667eea; text-decoration: none; font-size: 14px;">
										{{.ResetLink}}
									</a>
								</p>
							</div>
							<div style="margin: 30px 0; text-align: center;">
								<p style="margin: 0; color: #64748b; font-size: 14px; line-height: 1.6;">
									🔒 Link ini akan kedaluwarsa dalam 15 menit demi keamanan Anda.<br>
									Jika Anda tidak meminta reset password, abaikan email ini.
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