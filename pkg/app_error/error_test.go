package apperror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("should create new AppError with all fields", func(t *testing.T) {
		statusCode := 400
		code := "INVALID_REQUEST"
		message := "Invalid request data"
		details := "Field 'email' is required"

		err := New(statusCode, code, message, details)

		assert.NotNil(t, err)
		assert.Equal(t, statusCode, err.StatusCode)
		assert.Equal(t, code, err.Code)
		assert.Equal(t, message, err.Message)
		assert.Equal(t, details, err.Details)
	})

	t.Run("should create AppError without details", func(t *testing.T) {
		statusCode := 404
		code := "NOT_FOUND"
		message := "Resource not found"

		err := New(statusCode, code, message, "")

		assert.NotNil(t, err)
		assert.Equal(t, statusCode, err.StatusCode)
		assert.Equal(t, code, err.Code)
		assert.Equal(t, message, err.Message)
		assert.Empty(t, err.Details)
	})

	t.Run("should create different error types", func(t *testing.T) {
		testCases := []struct {
			name       string
			statusCode int
			code       string
			message    string
			details    string
		}{
			{
				name:       "Bad Request",
				statusCode: 400,
				code:       "BAD_REQUEST",
				message:    "Bad request error",
				details:    "Some details",
			},
			{
				name:       "Unauthorized",
				statusCode: 401,
				code:       "UNAUTHORIZED",
				message:    "Unauthorized access",
				details:    "Invalid credentials",
			},
			{
				name:       "Forbidden",
				statusCode: 403,
				code:       "FORBIDDEN",
				message:    "Access forbidden",
				details:    "Insufficient permissions",
			},
			{
				name:       "Not Found",
				statusCode: 404,
				code:       "NOT_FOUND",
				message:    "Resource not found",
				details:    "",
			},
			{
				name:       "Internal Server Error",
				statusCode: 500,
				code:       "INTERNAL_ERROR",
				message:    "Internal server error",
				details:    "Database connection failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := New(tc.statusCode, tc.code, tc.message, tc.details)

				assert.Equal(t, tc.statusCode, err.StatusCode)
				assert.Equal(t, tc.code, err.Code)
				assert.Equal(t, tc.message, err.Message)
				assert.Equal(t, tc.details, err.Details)
			})
		}
	})
}

func TestAppError_Error(t *testing.T) {
	t.Run("should return message as error string", func(t *testing.T) {
		message := "Test error message"
		err := New(500, "TEST_ERROR", message, "")

		assert.Equal(t, message, err.Error())
	})

	t.Run("should work with different messages", func(t *testing.T) {
		testMessages := []string{
			"User not found",
			"Invalid credentials",
			"Database connection failed",
			"Email already exists",
		}

		for _, msg := range testMessages {
			err := New(400, "TEST", msg, "")
			assert.Equal(t, msg, err.Error())
		}
	})
}

func TestAppError_AsError(t *testing.T) {
	t.Run("should implement error interface", func(t *testing.T) {
		var err error
		appErr := New(500, "TEST", "Test message", "")

		err = appErr
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Test message")
	})
}

func TestAppError_Fields(t *testing.T) {
	t.Run("should have correct status codes", func(t *testing.T) {
		statusCodes := []int{200, 201, 400, 401, 403, 404, 500, 503}

		for _, code := range statusCodes {
			err := New(code, "TEST", "Test", "")
			assert.Equal(t, code, err.StatusCode)
		}
	})

	t.Run("should handle empty strings", func(t *testing.T) {
		err := New(400, "", "", "")

		assert.Equal(t, 400, err.StatusCode)
		assert.Empty(t, err.Code)
		assert.Empty(t, err.Message)
		assert.Empty(t, err.Details)
	})

	t.Run("should handle special characters", func(t *testing.T) {
		specialMessage := "Error: Invalid input! @#$%^&*()"
		specialDetails := "Details with 特殊字符 and émojis 🚀"

		err := New(400, "SPECIAL_TEST", specialMessage, specialDetails)

		assert.Equal(t, specialMessage, err.Message)
		assert.Equal(t, specialDetails, err.Details)
	})
}

func TestAppError_CommonPatterns(t *testing.T) {
	t.Run("authentication error pattern", func(t *testing.T) {
		err := New(401, "INVALID_CREDENTIALS", "Email atau password salah", "")

		assert.Equal(t, 401, err.StatusCode)
		assert.Equal(t, "INVALID_CREDENTIALS", err.Code)
	})

	t.Run("validation error pattern", func(t *testing.T) {
		err := New(400, "VALIDATION_FAILED", "Data tidak valid", "Field 'email' harus diisi")

		assert.Equal(t, 400, err.StatusCode)
		assert.NotEmpty(t, err.Details)
	})

	t.Run("not found error pattern", func(t *testing.T) {
		err := New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "")

		assert.Equal(t, 404, err.StatusCode)
		assert.Empty(t, err.Details)
	})

	t.Run("server error pattern", func(t *testing.T) {
		err := New(500, "DATABASE_ERROR", "Terjadi kesalahan server", "Connection timeout")

		assert.Equal(t, 500, err.StatusCode)
		assert.NotEmpty(t, err.Details)
	})
}


func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(500, "TEST_ERROR", "Test error message", "Test details")
	}
}

func BenchmarkError(b *testing.B) {
	err := New(500, "TEST_ERROR", "Test error message", "Test details")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
