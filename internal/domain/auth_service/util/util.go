package util

import mainutils "pingspot/pkg/utils/main_util"

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
<html lang="id" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Verifikasi Akun PingSpot</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			Verifikasi email Anda untuk mulai menggunakan PingSpot. Link berlaku selama 5 menit.
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
									Selamat datang di PingSpot!
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
											✉️
										</td>
									</tr>
								</table>
	
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Halo {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 8px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									Terima kasih telah bergabung dengan PingSpot! Untuk mulai menggunakan dan mengamankan akun Anda, silakan verifikasi alamat email Anda.
								</p>
	
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 32px auto 18px;">
									<tr>
										<td align="center" bgcolor="#6C5CE7" style="background-color: #6C5CE7; border-radius: 12px; box-shadow: 0 6px 18px rgba(108, 92, 231, 0.35);">
											<a href="{{.VerificationLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												Verifikasi Akun Saya &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Expiry notice -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 28px;">
									<tr>
										<td style="padding: 6px 14px; background-color: #fffbeb; border-radius: 999px; color: #92400e; font-size: 13px; font-weight: 600;">
											⏱️ Berlaku selama 5 menit
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-left: 4px solid #6C5CE7; border-radius: 12px;">
									<tr>
										<td style="padding: 18px 20px;">
											<p style="margin: 0 0 6px; color: #3b3866; font-size: 14px; font-weight: 700;">
												Tombol tidak berfungsi?
											</p>
											<p style="margin: 0; color: #64688a; font-size: 13px; line-height: 1.5;">
												Salin dan tempel link ini ke browser Anda:
											</p>
											<p style="margin: 8px 0 0; word-break: break-all;">
												<a href="{{.VerificationLink}}" style="color: #6C5CE7; text-decoration: underline; font-size: 13px;">{{.VerificationLink}}</a>
											</p>
										</td>
									</tr>
								</table>
							</td>
						</tr>
	
						<!-- Security note -->
						<tr>
							<td style="padding: 8px 40px 40px;">
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%">
									<tr>
										<td style="border-top: 1px solid #ebe9f7; padding-top: 22px; text-align: center;">
											<p style="margin: 0; color: #64688a; font-size: 14px; line-height: 1.6;">
												🔒 Link ini akan kedaluwarsa dalam 5 menit demi keamanan Anda.<br>
												Jika Anda tidak membuat akun, abaikan email ini.
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
									<a href="andreanjabar18@gmail.com" style="color: #6C5CE7; text-decoration: none; font-weight: 600;">andreanjabar18@gmail.com</a>
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
<html lang="id" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Reset Password PingSpot</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			Kami menerima permintaan reset password untuk akun Anda. Link berlaku selama 15 menit.
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
									Reset Password Anda
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
											🔑
										</td>
									</tr>
								</table>
	
								<h2 style="margin: 0 0 14px; color: #1e1b3a; font-size: 24px; font-weight: 700; text-align: center;">
									Halo {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 8px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									Kami menerima permintaan untuk mereset password akun Anda. Klik tombol di bawah untuk melanjutkan proses reset password.
								</p>
	
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 32px auto 18px;">
									<tr>
										<td align="center" bgcolor="#6C5CE7" style="background-color: #6C5CE7; border-radius: 12px; box-shadow: 0 6px 18px rgba(108, 92, 231, 0.35);">
											<a href="{{.ResetLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												Reset Password &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Expiry notice -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 28px;">
									<tr>
										<td style="padding: 6px 14px; background-color: #fffbeb; border-radius: 999px; color: #92400e; font-size: 13px; font-weight: 600;">
											⏱️ Berlaku selama 15 menit
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-left: 4px solid #6C5CE7; border-radius: 12px;">
									<tr>
										<td style="padding: 18px 20px;">
											<p style="margin: 0 0 6px; color: #3b3866; font-size: 14px; font-weight: 700;">
												Tombol tidak berfungsi?
											</p>
											<p style="margin: 0; color: #64688a; font-size: 13px; line-height: 1.5;">
												Salin dan tempel link ini ke browser Anda:
											</p>
											<p style="margin: 8px 0 0; word-break: break-all;">
												<a href="{{.ResetLink}}" style="color: #6C5CE7; text-decoration: underline; font-size: 13px;">{{.ResetLink}}</a>
											</p>
										</td>
									</tr>
								</table>
							</td>
						</tr>
	
						<!-- Security note -->
						<tr>
							<td style="padding: 8px 40px 40px;">
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%">
									<tr>
										<td style="border-top: 1px solid #ebe9f7; padding-top: 22px; text-align: center;">
											<p style="margin: 0; color: #64688a; font-size: 14px; line-height: 1.6;">
												🔒 Link ini akan kedaluwarsa dalam 15 menit demi keamanan Anda.<br>
												Jika Anda tidak meminta reset password, abaikan email ini &mdash; password Anda tidak akan berubah.
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
									<a href="andreanjabar18@gmail.com" style="color: #6C5CE7; text-decoration: none; font-weight: 600;">andreanjabar18@gmail.com</a>
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