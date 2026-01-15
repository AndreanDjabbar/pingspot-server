package mainutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"pingspot/pkg/utils/env"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/gomail.v2"
)

func GetClientIP(c *fiber.Ctx) string {
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		if idx := len(ip); idx > 0 {
			for i := 0; i < idx; i++ {
				if ip[i] == ',' {
					return ip[:i]
				}
			}
			return ip
		}
	}

	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	return c.IP()
}

func GetHTTPClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := len(ip); idx > 0 {
			for i := 0; i < idx; i++ {
				if ip[i] == ',' {
					return ip[:i]
				}
			}
			return ip
		}
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}

func GetHTTPUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

func GetDeviceInfo(c *fiber.Ctx) string {
	return fmt.Sprintf("IP: %s, UA: %s", GetClientIP(c), GetUserAgent(c))
}

func GetKeyPath(filename string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "keys", filename)
}

type EmailType string

const (
	EmailTypeVerification     EmailType = "verification"
	EmailTypePasswordReset    EmailType = "password_reset"
	EmailTypeProgressReminder EmailType = "progress_reminder"
)

type EmailData struct {
	To            string
	Subject       string
	RecipientName string
	EmailType     EmailType
	BodyTempate   string
	TemplateData  map[string]any
}

func SendEmail(data EmailData) error {
	if data.To == "" || data.RecipientName == "" {
		return fmt.Errorf("recipient email and name cannot be empty")
	}

	email := env.EmailEmail()
	password := env.EmailPassword()

	if email == "" || password == "" {
		return fmt.Errorf("email credentials not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", data.To)
	m.SetHeader("Subject", data.Subject)

	body, err := RenderEmailTemplate(data)

	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}
	m.SetBody("text/html", body)
	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func RenderEmailTemplate(data EmailData) (string, error) {
	var templateHTML string
	var templateData any

	switch data.EmailType {
	case EmailTypeVerification:
		verificationLink, ok := data.TemplateData["VerificationLink"].(string)
		if !ok || verificationLink == "" {
			return "", fmt.Errorf("verification link is required for verification email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName         string
			VerificationLink string
		}{
			UserName:         data.RecipientName,
			VerificationLink: verificationLink,
		}

	case EmailTypePasswordReset:
		resetLink, ok := data.TemplateData["ResetLink"].(string)
		if !ok || resetLink == "" {
			return "", fmt.Errorf("reset link is required for password reset email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName  string
			ResetLink string
		}{
			UserName:  data.RecipientName,
			ResetLink: resetLink,
		}

	case EmailTypeProgressReminder:
		reportTitle, ok := data.TemplateData["ReportTitle"].(string)
		if !ok || reportTitle == "" {
			return "", fmt.Errorf("report title is required for progress reminder email")
		}
		reportLink, ok := data.TemplateData["ReportLink"].(string)
		if !ok || reportLink == "" {
			return "", fmt.Errorf("report link is required for progress reminder email")
		}
		daysRemaining, ok := data.TemplateData["DaysRemaining"].(int)
		if !ok {
			return "", fmt.Errorf("days remaining is required for progress reminder email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName      string
			ReportTitle   string
			ReportLink    string
			DaysRemaining int
		}{
			UserName:      data.RecipientName,
			ReportTitle:   reportTitle,
			ReportLink:    reportLink,
			DaysRemaining: daysRemaining,
		}

	default:
		return "", fmt.Errorf("unsupported email type: %s", data.EmailType)
	}

	tmpl, err := template.New(string(data.EmailType)).Parse(templateHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func StrPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func BoolPtrOrNil(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

func IntPtrOrNil(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func Int64PtrOrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func StringToInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int: %w", err)
	}
	return value, nil
}

func StringToTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	layout := time.RFC3339
	value, err := time.Parse(layout, s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to time.Time: %w", err)
	}
	return &value, nil
}

func StringToFloat64(s string) (float64, error) {
	if s == "" {
		return 0.0, nil
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to convert string to float64: %w", err)
	}

	return value, nil
}

func StringPtrToObjectIDPtr(s *string) (*primitive.ObjectID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	objectID, err := primitive.ObjectIDFromHex(*s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to ObjectID: %w", err)
	}
	return &objectID, nil
}

func ObjectIDPtrToStringPtr(id *primitive.ObjectID) *string {
	if id == nil {
		return nil
	}
	s := id.Hex()
	return &s
}

func StringToBool(s string) (*bool, error) {
	if s == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to bool: %w", err)
	}
	return &value, nil
}

func StringToUint(s string) (uint, error) {
	if s == "" {
		return 0, nil
	}
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to uint: %w", err)
	}
	return uint(value), nil
}

func MustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
