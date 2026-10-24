package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"pingspot/internal/domain/auth_service/dto"
	userRepo "pingspot/internal/domain/user_service/repository"
	redisRepo "pingspot/internal/infrastructure/redis"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"pingspot/internal/util"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func sendVerificationEmail(to, username, verificationLink string) error {
	return util.SendEmail(util.EmailData{
		To:            to,
		Subject:       "Pingspot Account Verification",
		RecipientName: username,
		BodyTempate:   getVerificationEmailTemplate(),
		EmailType:     util.EmailTypeVerification,
		TemplateData: map[string]interface{}{
			"VerificationLink": verificationLink,
		},
	})
}

func getVerificationEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Verify Your PingSpot Account</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			Verify your email to start using PingSpot. The link is valid for 5 minutes.
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
									Welcome to PingSpot!
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
									Hello {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 8px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									Thank you for joining PingSpot! To start using and secure your account, please verify your email address.
								</p>
	
								<!-- CTA -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 32px auto 18px;">
									<tr>
										<td align="center" bgcolor="#6C5CE7" style="background-color: #6C5CE7; border-radius: 12px; box-shadow: 0 6px 18px rgba(108, 92, 231, 0.35);">
											<a href="{{.VerificationLink}}" target="_blank"
											style="display: inline-block; padding: 16px 36px; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 16px; border-radius: 12px;">
												Verify My Account &rarr;
											</a>
										</td>
									</tr>
								</table>
	
								<!-- Expiry notice -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin: 0 auto 28px;">
									<tr>
										<td style="padding: 6px 14px; background-color: #fffbeb; border-radius: 999px; color: #92400e; font-size: 13px; font-weight: 600;">
											⏱️ Valid for 5 minutes
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-left: 4px solid #6C5CE7; border-radius: 12px;">
									<tr>
										<td style="padding: 18px 20px;">
											<p style="margin: 0 0 6px; color: #3b3866; font-size: 14px; font-weight: 700;">
												Button not working?
											</p>
											<p style="margin: 0; color: #64688a; font-size: 13px; line-height: 1.5;">
												Copy and paste this link into your browser:
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
												🔒 This link will expire in 5 minutes for your security.<br>
												If you didn't create an account, you can ignore this email.
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

func sendPasswordResetEmail(to, username, resetLink string, isDisabled bool) error {
	return util.SendEmail(util.EmailData{
		To:            to,
		Subject:       "Pingspot Password Reset",
		RecipientName: username,
		BodyTempate:   getPasswordResetEmailTemplate(),
		EmailType:     util.EmailTypePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetLink": resetLink,
		},
		DisabledNotifications: isDisabled,
	})
}

func getPasswordResetEmailTemplate() string {
	return `<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta name="color-scheme" content="light">
		<meta name="supported-color-schemes" content="light">
		<title>Reset Your PingSpot Password</title>
	</head>
	<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, sans-serif; background-color: #f4f3fb; line-height: 1.6; -webkit-text-size-adjust: 100%;">
	
		<!-- Preheader (preview text in inbox) -->
		<div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: #f4f3fb; font-size: 1px; line-height: 1px;">
			We received a request to reset your account password. The link is valid for 5 minutes.
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
									Reset Your Password
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
									Hello {{.UserName}}! 👋
								</h2>
								<p style="margin: 0 0 8px; color: #4b5068; font-size: 16px; text-align: center; line-height: 1.7;">
									We received a request to reset your account password. Click the button below to continue the reset process.
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
											⏱️ Valid for 5 minutes
										</td>
									</tr>
								</table>
	
								<!-- Link fallback -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%" style="background-color: #f1effd; border: 1px solid #ddd9fa; border-left: 4px solid #6C5CE7; border-radius: 12px;">
									<tr>
										<td style="padding: 18px 20px;">
											<p style="margin: 0 0 6px; color: #3b3866; font-size: 14px; font-weight: 700;">
												Button not working?
											</p>
											<p style="margin: 0; color: #64688a; font-size: 13px; line-height: 1.5;">
												Copy and paste this link into your browser:
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
												🔒 This link will expire in 5 minutes for your security.<br>
												If you didn't request a password reset, you can ignore this email &mdash; your password will not change.
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

func getRefreshTokenDuration() time.Duration {
	ageStr := util.RefreshTokenAge()
	age, err := strconv.Atoi(ageStr)
	if err != nil || age <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(age) * time.Second
}

type AuthService struct {
	db              *gorm.DB
	userRepo        userRepo.UserRepository
	userSessionRepo userRepo.UserSessionRepository
	userProfileRepo userRepo.UserProfileRepository
	redisRepo       redisRepo.CacheRepository
}

func NewAuthService(
	db *gorm.DB,
	userRepo userRepo.UserRepository,
	userProfileRepo userRepo.UserProfileRepository,
	userSessionRepo userRepo.UserSessionRepository,
	redisRepo redisRepo.CacheRepository,
) *AuthService {
	return &AuthService{
		db:              db,
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userSessionRepo: userSessionRepo,
		redisRepo:       redisRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, isVerified bool) (*model.User, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("Registering new user",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
		zap.String("provider", req.Provider),
	)

	tx := s.db.Begin()
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

	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		tx.Rollback()
		return nil, util.New(400, "EMAIL_ALREADY_REGISTERED", "Email already registered", "", nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, util.New(500, "USER_CHECK_FAILED", "Failed to check user data", err.Error(), nil)
	}

	hashedPassword, err := util.HashString(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, util.New(500, "PASSWORD_HASH_FAILED", "Failed to hash password", err.Error(), nil)
	}

	user := model.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   &hashedPassword,
		FullName:   req.FullName,
		Provider:   model.Provider(req.Provider),
		ProviderID: req.ProviderID,
		IsVerified: isVerified,
	}

	createdUser, err := s.userRepo.CreateTX(ctx, tx, &user)
	if err != nil {
		tx.Rollback()
		return nil, util.New(500, "USER_CREATE_FAILED", "Failed to create user", err.Error(), nil)
	}

	newProfile := model.UserProfile{
		UserID: createdUser.ID,
	}

	if _, err := s.userProfileRepo.CreateTX(ctx, tx, &newProfile); err != nil {
		tx.Rollback()
		return nil, util.New(500, "PROFILE_CREATE_FAILED", "Failed to create user profile", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to save changes", err.Error(), nil)
	}

	logger.Info("User registered successfully",
		zap.String("request_id", requestID),
		zap.Uint("user_id", createdUser.ID),
		zap.String("email", createdUser.Email),
	)

	return createdUser, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*model.User, string, string, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("User login attempt",
		zap.String("request_id", requestID),
		zap.String("email or username", req.EmailOrUsername),
	)

	user, err := s.userRepo.GetByEmailOrUsername(ctx, req.EmailOrUsername)
	if err != nil {
		return nil, "", "", util.New(401, "INVALID_CREDENTIALS", "Incorrect email or password", err.Error(), nil)
	}

	if model.Provider(req.Provider) == model.ProviderEmail {
		if !util.CheckHashString(req.Password, *user.Password) {
			return nil, "", "", util.New(401, "INVALID_CREDENTIALS", "Incorrect email or password", "", nil)
		}
	}

	if !user.IsVerified {
		randomCode1, err := util.GenerateRandomCode(150)
		if err != nil {
			logger.Error("Failed to generate random code",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", util.New(500, "CODE_GENERATION_FAILED", "Failed to generate random code", err.Error(), nil)
		}
		randomCode2, err := util.GenerateRandomCode(150)
		if err != nil {
			logger.Error("Failed to generate random code",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", util.New(500, "CODE_GENERATION_FAILED", "Failed to generate random code", err.Error(), nil)
		}
		verificationLink := fmt.Sprintf("%s/verify-account/%s/%d/%s", util.ClientURL(), randomCode1, user.ID, randomCode2)

		linkData := map[string]string{
			"link1": randomCode1,
			"link2": randomCode2,
		}
		linkJSON, err := json.Marshal(linkData)
		if err != nil {
			logger.Error("Failed to marshal verification link data",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", util.New(500, "VERIFICATION_CODE_SAVE_FAILED", "Failed to save verification code", err.Error(), nil)
		}
		redisKey := fmt.Sprintf("link:%d", user.ID)
		err = s.redisRepo.Set(context.Background(), redisKey, linkJSON, 5*time.Minute)
		if err != nil {
			logger.Error("Failed to save verification link to Redis",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", util.New(500, "VERIFICATION_CODE_REDIS_FAILED", "Failed to save verification code to Redis", err.Error(), nil)
		}
		go sendVerificationEmail(user.Email, user.Username, verificationLink)
		return nil, "", "", util.New(403, "ACCOUNT_NOT_VERIFIED", "Account is not verified. Please check your email to verify it", "", nil)
	}

	refreshTokenID := uuid.New().String()
	refreshToken := util.GenerateRefreshToken(user.ID, refreshTokenID)
	hashedRefreshToken := util.HashSHA256String(refreshToken)

	tx := s.db.Begin()
	userSession, err := s.userSessionRepo.CreateTX(ctx, tx, &model.UserSession{
		UserID:             user.ID,
		ExpiresAt:          time.Now().Add(getRefreshTokenDuration()).Unix(),
		IsActive:           true,
		RefreshTokenID:     refreshTokenID,
		HashedRefreshToken: hashedRefreshToken,
		IPAddress:          req.IPAddress,
		UserAgent:          req.UserAgent,
	})
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to create user session",
			zap.String("request_id", requestID),
			zap.Uint("user_id", user.ID),
			zap.Error(err),
		)
		return nil, "", "", util.New(500, "USER_SESSION_CREATE_FAILED", "Failed to create user session", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit user session transaction",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", util.New(500, "USER_SESSION_COMMIT_FAILED", "Failed to save user session", err.Error(), nil)
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
	err = s.redisRepo.Set(context.Background(), refreshKey, hashedRefreshToken, getRefreshTokenDuration())
	if err != nil {
		logger.Error("Failed to save refresh token to Redis",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", util.New(500, "REFRESH_TOKEN_SAVE_FAILED", "Failed to save refresh token", err.Error(), nil)
	}

	userSessionKey := fmt.Sprintf("user_session:%d", user.ID)
	err = s.redisRepo.SAdd(context.Background(), userSessionKey, userSession.ID)
	if err != nil {
		logger.Error("Failed to save user session ID to Redis set",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", util.New(500, "USER_SESSION_SAVE_FAILED", "Failed to save user session", err.Error(), nil)
	}

	sessionKey := fmt.Sprintf("session:%d", userSession.ID)
	err = s.redisRepo.Set(context.Background(), sessionKey, user.ID, getRefreshTokenDuration())
	if err != nil {
		logger.Error("Failed to save session data to Redis",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", util.New(500, "SESSION_SAVE_FAILED", "Failed to save session data", err.Error(), nil)
	}

	accessToken := util.GenerateAccessToken(user.ID, userSession.ID, user.Email, user.Username, user.FullName)

	logger.Info("User logged in successfully",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := util.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", util.New(401, "INVALID_REFRESH_TOKEN", "Invalid refresh token", err.Error(), nil)
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", "", util.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Invalid user_id claim", "", nil)
	}

	refreshTokenID, ok := claims["refresh_token_id"].(string)
	if !ok || refreshTokenID == "" {
		return "", "", util.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Invalid refresh_token_id claim", "", nil)
	}

	userID := uint(userIDFloat)
	hashedRefreshToken := util.HashSHA256String(refreshToken)
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)

	storedHashedRefreshToken, err := s.redisRepo.Get(ctx, refreshKey)
	if err != nil {
		logger.Warn("Failed to get refresh token from Redis, fallback to PostgreSQL",
			zap.String("refresh_token_id", refreshTokenID),
			zap.Error(err),
		)
	} else {
		if storedHashedRefreshToken != hashedRefreshToken {
			return "", "", util.New(401, "REFRESH_TOKEN_INVALID", "Refresh token does not match", "", nil)
		}
	}

	userSession, err := s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		return "", "", util.New(401, "USER_SESSION_NOT_FOUND", "User session not found", err.Error(), nil)
	}

	if !userSession.IsActive {
		return "", "", util.New(401, "SESSION_INACTIVE", "Session is inactive", "", nil)
	}

	if time.Now().Unix() > userSession.ExpiresAt {
		return "", "", util.New(401, "SESSION_EXPIRED", "Session has expired", "", nil)
	}

	if userSession.HashedRefreshToken != hashedRefreshToken {
		return "", "", util.New(401, "REFRESH_TOKEN_INVALID", "Refresh token does not match", "", nil)
	}

	if userSession.UserID != userID {
		return "", "", util.New(401, "SESSION_USER_MISMATCH", "Session does not match the user", "", nil)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", util.New(404, "USER_NOT_FOUND", "User not found", err.Error(), nil)
		}
		return "", "", util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	refreshDuration := getRefreshTokenDuration()
	newExpiresAt := time.Now().Add(refreshDuration).Unix()

	newRefreshTokenID := uuid.New().String()
	newRefreshToken := util.GenerateRefreshToken(userID, newRefreshTokenID)
	newHashedRefreshToken := util.HashSHA256String(newRefreshToken)

	oldRefreshKey := refreshKey
	newRefreshKey := fmt.Sprintf("refresh_token:%s", newRefreshTokenID)
	sessionKey := fmt.Sprintf("session:%d", userSession.ID)

	userSession.RefreshTokenID = newRefreshTokenID
	userSession.HashedRefreshToken = newHashedRefreshToken
	userSession.ExpiresAt = newExpiresAt

	tx := s.db.Begin()
	if tx.Error != nil {
		return "", "", util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}

	if err := s.userSessionRepo.UpdateTX(ctx, tx, userSession); err != nil {
		tx.Rollback()
		return "", "", util.New(500, "SESSION_UPDATE_FAILED", "Failed to update session", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", "", util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to commit transaction", err.Error(), nil)
	}

	if err := s.redisRepo.Set(ctx, newRefreshKey, newHashedRefreshToken, refreshDuration); err != nil {
		logger.Warn("Failed to save new refresh token to Redis",
			zap.String("refresh_token_id", newRefreshTokenID),
			zap.Error(err),
		)
	}

	if err := s.redisRepo.Set(ctx, sessionKey, userID, refreshDuration); err != nil {
		logger.Warn("Failed to save session to Redis",
			zap.Uint("session_id", userSession.ID),
			zap.Error(err),
		)
	}

	if err := s.redisRepo.Del(ctx, oldRefreshKey); err != nil {
		logger.Warn("Failed to delete old refresh token from Redis",
			zap.String("refresh_token_id", refreshTokenID),
			zap.Error(err),
		)
	}

	accessToken := util.GenerateAccessToken(
		userID,
		userSession.ID,
		user.Email,
		user.Username,
		user.FullName,
	)

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenClaims, err := util.ValidateRefreshToken(refreshToken)
	if err != nil {
		logger.Error("Failed to validate refresh token", zap.Error(err))
		return util.New(401, "INVALID_REFRESH_TOKEN", "Invalid refresh token", err.Error(), nil)
	}

	refreshTokenID, ok := refreshTokenClaims["refresh_token_id"].(string)
	if !ok || refreshTokenID == "" {
		return util.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Invalid refresh_token_id claim", "", nil)
	}

	userIDFloat, ok := refreshTokenClaims["user_id"].(float64)
	if !ok {
		return util.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Invalid user_id claim", "", nil)
	}
	userID := uint(userIDFloat)

	userSession, err := s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.New(401, "USER_SESSION_NOT_FOUND", "User session not found", err.Error(), nil)
		}
		return util.New(500, "USER_SESSION_FETCH_FAILED", "Failed to retrieve user session", err.Error(), nil)
	}

	if userSession.UserID != userID {
		return util.New(401, "SESSION_USER_MISMATCH", "Session does not match the user", "", nil)
	}

	if !userSession.IsActive {
		refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
		sessionKey := fmt.Sprintf("session:%d", userSession.ID)
		userSessionKey := fmt.Sprintf("user_session:%d", userID)

		if err := s.redisRepo.Del(ctx, refreshKey); err != nil {
			logger.Warn("Failed to delete refresh token from Redis", zap.Error(err))
		}
		if err := s.redisRepo.Del(ctx, sessionKey); err != nil {
			logger.Warn("Failed to delete session data from Redis", zap.Error(err))
		}
		if err := s.redisRepo.SRem(ctx, userSessionKey, userSession.ID); err != nil {
			logger.Warn("Failed to remove user session ID from Redis set", zap.Error(err))
		}

		return nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return util.New(500, "TRANSACTION_START_FAILED", "Failed to start transaction", tx.Error.Error(), nil)
	}

	userSession.IsActive = false
	if err := s.userSessionRepo.UpdateTX(ctx, tx, userSession); err != nil {
		tx.Rollback()
		return util.New(500, "USER_SESSION_UPDATE_FAILED", "Failed to update user session", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.New(500, "TRANSACTION_COMMIT_FAILED", "Failed to commit transaction", err.Error(), nil)
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
	sessionKey := fmt.Sprintf("session:%d", userSession.ID)
	userSessionKey := fmt.Sprintf("user_session:%d", userID)

	if err := s.redisRepo.Del(ctx, refreshKey); err != nil {
		logger.Warn("Failed to delete refresh token from Redis", zap.Error(err))
	}

	if err := s.redisRepo.Del(ctx, sessionKey); err != nil {
		logger.Warn("Failed to delete session data from Redis", zap.Error(err))
	}

	if err := s.redisRepo.SRem(ctx, userSessionKey, userSession.ID); err != nil {
		logger.Warn("Failed to remove user session ID from Redis set", zap.Error(err))
	}

	return nil
}

func (s *AuthService) VerifyUser(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "USER_NOT_FOUND", "User not found", err.Error(), nil)
		}
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	if user.IsVerified {
		return nil, util.New(400, "ALREADY_VERIFIED", "Account is already verified", "", nil)
	}

	user.IsVerified = true

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, util.New(500, "USER_SAVE_FAILED", "Failed to save user data", err.Error(), nil)
	}

	return user, nil
}

func (s *AuthService) UpdateUserByEmail(ctx context.Context, email string, updatedUser *model.User) (*model.User, error) {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.New(404, "USER_NOT_FOUND", "User not found", err.Error(), nil)
		}
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to find user", err.Error(), nil)
	}

	user, err := s.userRepo.UpdateByEmail(ctx, email, updatedUser)
	if err != nil {
		return nil, util.New(500, "USER_UPDATE_FAILED", "Failed to update user", err.Error(), nil)
	}

	return user, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}
	return user, nil
}

func (s *AuthService) ForgotPasswordEmailVerification(ctx context.Context, req dto.ForgotPasswordEmailVerificationRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.New(404, "USER_NOT_FOUND", "User not found", err.Error(), nil)
		}
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	if user != nil {
		existingRedisKey := fmt.Sprintf("forgot_password:%s", req.Email)
		remainingTime, err := s.redisRepo.TTL(ctx, existingRedisKey)
		if err == nil && remainingTime > 0 {
			return util.New(
				429,
				"TOO_MANY_REQUESTS",
				fmt.Sprintf("You have already made a request. Please check your email or try again in %d seconds", int(remainingTime.Seconds())),
				"",
				map[string]any{
					"retry_after_seconds": int(remainingTime.Seconds()),
				},
			)
		}

		verificationCode, err := util.GenerateRandomCode(200)
		if err != nil {
			logger.Error("Failed to generate verification code", zap.Error(err))
			return util.New(500, "CODE_GENERATION_FAILED", "Failed to generate verification code", err.Error(), nil)
		}

		redisKey := fmt.Sprintf("forgot_password:%s", req.Email)
		err = s.redisRepo.Set(ctx, redisKey, verificationCode, 300*time.Second)
		if err != nil {
			logger.Error("Failed to save verification code to Redis", zap.Error(err))
			return util.New(500, "CODE_GENERATION_FAILED", "Failed to generate verification code", err.Error(), nil)
		}

		verificationLink := fmt.Sprintf("%s/forgot-password/verification?code=%s&email=%s", util.ClientURL(), verificationCode, req.Email)
		go sendPasswordResetEmail(req.Email, req.Email, verificationLink, user.IsDisableEmailNotification)
		return nil
	}
	return util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
}

func (s *AuthService) SendRegistrationVerificationEmail(ctx context.Context, user *model.User) error {
	randomCode1, err := util.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code 1", zap.Error(err))
		return util.New(500, "CODE_GENERATION_FAILED", "Failed to generate random code", err.Error(), nil)
	}
	randomCode2, err := util.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code 2", zap.Error(err))
		return util.New(500, "CODE_GENERATION_FAILED", "Failed to generate random code", err.Error(), nil)
	}

	verificationLink := fmt.Sprintf("%s/verification?code1=%s&userId=%d&code2=%s", util.ClientURL(), randomCode1, user.ID, randomCode2)

	linkData := map[string]string{
		"link1": randomCode1,
		"link2": randomCode2,
	}
	linkJSON, err := json.Marshal(linkData)
	if err != nil {
		logger.Error("Failed to marshal verification link data", zap.Error(err))
		return util.New(500, "MARSHAL_FAILED", "Failed to save verification code", err.Error(), nil)
	}

	redisKey := fmt.Sprintf("link:%d", user.ID)
	err = s.redisRepo.Set(ctx, redisKey, linkJSON, 300*time.Second)
	if err != nil {
		logger.Error("Failed to save verification link to Redis", zap.Error(err))
		return util.New(500, "REDIS_SAVE_FAILED", "Failed to save verification code to Redis", err.Error(), nil)
	}

	go sendVerificationEmail(user.Email, user.Username, verificationLink)

	return nil
}

func (s *AuthService) VerifyRegistrationCode(ctx context.Context, code1, code2 string, userID uint) (*model.User, error) {
	redisKey := fmt.Sprintf("link:%d", userID)
	linkData, err := s.redisRepo.Get(ctx, redisKey)
	if err != nil {
		logger.Error("Failed to get verification link from Redis", zap.Error(err))
		return nil, util.New(500, "REDIS_GET_FAILED", "Failed to retrieve verification code", err.Error(), nil)
	}

	var link map[string]string
	if err := json.Unmarshal([]byte(linkData), &link); err != nil {
		logger.Error("Failed to unmarshal verification link data", zap.Error(err))
		return nil, util.New(500, "UNMARSHAL_FAILED", "Failed to process verification link", err.Error(), nil)
	}

	if link["link1"] != code1 || link["link2"] != code2 {
		return nil, util.New(400, "INVALID_CODE", "Invalid verification link", "", nil)
	}

	user, err := s.VerifyUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.redisRepo.Del(ctx, redisKey); err != nil {
		logger.Warn("Failed to delete verification link from Redis", zap.Error(err))
	}

	return user, nil
}

func (s *AuthService) VerifyForgotPasswordCode(ctx context.Context, code, email string) error {
	redisKey := fmt.Sprintf("forgot_password:%s", email)
	storedCode, err := s.redisRepo.Get(ctx, redisKey)
	if err != nil {
		logger.Error("Failed to get verification code from Redis", zap.Error(err))
		return util.New(500, "REDIS_GET_FAILED", "Failed to retrieve verification code", err.Error(), nil)
	}

	if storedCode != code {
		return util.New(400, "INVALID_CODE", "Invalid verification link", "", nil)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword string) error {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		return err
	}
	if user == nil {
		return util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
	}

	hashedPassword, err := util.HashString(newPassword)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return util.New(500, "PASSWORD_HASH_FAILED", "Failed to hash the new password", err.Error(), nil)
	}

	user.Password = &hashedPassword

	_, err = s.UpdateUserByEmail(ctx, email, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, oauthEmail, oauthFullName, oauthGivenName, oauthName, oauthProviderID, provider, userIP, userAgent string) (string, string, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("Processing OAuth callback",
		zap.String("request_id", requestID),
		zap.String("provider", provider),
		zap.String("email", oauthEmail),
	)

	existingUser, err := s.GetUserByEmail(ctx, oauthEmail)
	if err != nil {
		logger.Error("Error retrieving user by email", zap.String("request_id", requestID), zap.Error(err))
		return "", "", util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	randomCode, err := util.GenerateRandomCode(5)
	if err != nil {
		logger.Error("Error generating random code", zap.String("request_id", requestID), zap.Error(err))
		return "", "", util.New(500, "CODE_GENERATION_FAILED", "Failed to generate random code", err.Error(), nil)
	}

	if existingUser == nil {
		nickName := oauthGivenName
		if nickName == "" {
			nickName = oauthName
		}

		newUser := dto.RegisterRequest{
			Username:   fmt.Sprintf("%s_%s", nickName, randomCode),
			Email:      oauthEmail,
			FullName:   oauthFullName,
			Provider:   strings.ToUpper(provider),
			ProviderID: &oauthProviderID,
		}
		createdUser, err := s.Register(ctx, newUser, true)
		if err != nil {
			logger.Error("Error registering new user from OAuth", zap.String("request_id", requestID), zap.String("provider", provider), zap.Error(err))
			return "", "", util.New(500, "USER_REGISTRATION_FAILED", "Failed to register new user", err.Error(), nil)
		}
		logger.Info("New user registered via OAuth", zap.String("request_id", requestID), zap.String("provider", provider), zap.Uint("user_id", createdUser.ID))
		existingUser = createdUser
	}

	loginReq := dto.LoginRequest{
		EmailOrUsername: existingUser.Email,
		Password:        "",
		IPAddress:       userIP,
		UserAgent:       userAgent,
		Provider:        strings.ToUpper(provider),
	}

	_, accessToken, refreshToken, err := s.Login(ctx, loginReq)
	if err != nil {
		logger.Error("Login failed for OAuth user", zap.String("request_id", requestID), zap.String("provider", provider), zap.Error(err))
		return "", "", util.New(500, "LOGIN_FAILED", "Failed to log in with the OAuth account", err.Error(), nil)
	}

	logger.Info("OAuth user logged in successfully", zap.String("request_id", requestID), zap.String("provider", provider), zap.Uint("user_id", existingUser.ID))

	return accessToken, refreshToken, nil
}
