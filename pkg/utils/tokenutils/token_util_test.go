package tokenutils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashString(t *testing.T) {
	t.Run("should hash string successfully", func(t *testing.T) {
		password := "mySecretPassword123"
		hash, err := HashString(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
	})

	t.Run("should generate different hashes for same input", func(t *testing.T) {
		password := "mySecretPassword123"
		hash1, err1 := HashString(password)
		hash2, err2 := HashString(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "bcrypt should generate different salts")
	})
}

func TestCheckHashString(t *testing.T) {
	t.Run("should validate correct password", func(t *testing.T) {
		password := "mySecretPassword123"
		hash, err := HashString(password)
		require.NoError(t, err)

		result := CheckHashString(password, hash)
		assert.True(t, result)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := "mySecretPassword123"
		wrongPassword := "wrongPassword"
		hash, err := HashString(password)
		require.NoError(t, err)

		result := CheckHashString(wrongPassword, hash)
		assert.False(t, result)
	})

	t.Run("should reject invalid hash", func(t *testing.T) {
		password := "mySecretPassword123"
		invalidHash := "invalid_hash_string"

		result := CheckHashString(password, invalidHash)
		assert.False(t, result)
	})
}

func TestHashSHA256String(t *testing.T) {
	t.Run("should generate consistent SHA256 hash", func(t *testing.T) {
		input := "testString"
		hash1 := HashSHA256String(input)
		hash2 := HashSHA256String(input)

		assert.Equal(t, hash1, hash2)
		assert.Len(t, hash1, 64)
	})

	t.Run("should generate different hashes for different inputs", func(t *testing.T) {
		input1 := "testString1"
		input2 := "testString2"

		hash1 := HashSHA256String(input1)
		hash2 := HashSHA256String(input2)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("should handle empty string", func(t *testing.T) {
		hash := HashSHA256String("")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})
}

func TestCheckHashSHA256String(t *testing.T) {
	t.Run("should validate correct SHA256 hash", func(t *testing.T) {
		input := "testString"
		hash := HashSHA256String(input)

		result := CheckHashSHA256String(input, hash)
		assert.True(t, result)
	})

	t.Run("should reject incorrect hash", func(t *testing.T) {
		input := "testString"
		wrongHash := "invalidhash123"

		result := CheckHashSHA256String(input, wrongHash)
		assert.False(t, result)
	})
}

func TestGenerateRandomCode(t *testing.T) {
	t.Run("should generate code with correct length", func(t *testing.T) {
		length := 10
		code, err := GenerateRandomCode(length)

		require.NoError(t, err)
		assert.Len(t, code, length)
	})

	t.Run("should generate different codes", func(t *testing.T) {
		code1, err1 := GenerateRandomCode(10)
		code2, err2 := GenerateRandomCode(10)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, code1, code2)
	})

	t.Run("should return error for invalid length", func(t *testing.T) {
		code, err := GenerateRandomCode(0)

		assert.Error(t, err)
		assert.Empty(t, code)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})

	t.Run("should return error for negative length", func(t *testing.T) {
		code, err := GenerateRandomCode(-5)

		assert.Error(t, err)
		assert.Empty(t, code)
	})

	t.Run("should only contain valid characters", func(t *testing.T) {
		code, err := GenerateRandomCode(100)
		require.NoError(t, err)

		validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
		for _, char := range code {
			assert.Contains(t, validChars, string(char))
		}
	})
}

func TestGenerateJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := uint(123)
	email := "test@example.com"
	username := "testuser"
	fullName := "Test User"

	t.Run("should generate valid JWT", func(t *testing.T) {
		token, err := GenerateJWT(userID, secret, email, username, fullName)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("should return error for empty secret", func(t *testing.T) {
		token, err := GenerateJWT(userID, []byte(""), email, username, fullName)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT secret cannot be empty")
	})

	t.Run("should include correct claims", func(t *testing.T) {
		token, err := GenerateJWT(userID, secret, email, username, fullName)
		require.NoError(t, err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		require.NoError(t, err)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)

		assert.Equal(t, float64(userID), claims["user_id"])
		assert.Equal(t, email, claims["email"])
		assert.Equal(t, username, claims["username"])
		assert.Equal(t, fullName, claims["full_name"])
		assert.NotNil(t, claims["exp"])
		assert.NotNil(t, claims["iat"])
	})

	t.Run("should set expiration time correctly", func(t *testing.T) {
		token, err := GenerateJWT(userID, secret, email, username, fullName)
		require.NoError(t, err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		require.NoError(t, err)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)

		exp, ok := claims["exp"].(float64)
		require.True(t, ok)

		expectedExp := time.Now().Add(time.Hour * 10).Unix()
		assert.InDelta(t, expectedExp, exp, 5)
	})
}

func TestParseJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := uint(123)
	email := "test@example.com"
	username := "testuser"
	fullName := "Test User"

	t.Run("should parse valid JWT", func(t *testing.T) {
		token, err := GenerateJWT(userID, secret, email, username, fullName)
		require.NoError(t, err)

		claims, err := ParseJWT(token, secret)
		require.NoError(t, err)

		assert.Equal(t, float64(userID), claims["user_id"])
		assert.Equal(t, email, claims["email"])
		assert.Equal(t, username, claims["username"])
		assert.Equal(t, fullName, claims["full_name"])
	})

	t.Run("should return error for invalid token", func(t *testing.T) {
		claims, err := ParseJWT("invalid.token.string", secret)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should return error for empty secret", func(t *testing.T) {
		token, _ := GenerateJWT(userID, secret, email, username, fullName)

		claims, err := ParseJWT(token, []byte(""))

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "JWT secret cannot be empty")
	})

	t.Run("should return error for wrong secret", func(t *testing.T) {
		token, err := GenerateJWT(userID, secret, email, username, fullName)
		require.NoError(t, err)

		wrongSecret := []byte("wrong-secret-key")
		claims, err := ParseJWT(token, wrongSecret)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should return error for expired token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id":   userID,
			"email":     email,
			"username":  username,
			"full_name": fullName,
			"exp":       time.Now().Add(-1 * time.Hour).Unix(),
			"iat":       time.Now().Add(-2 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		require.NoError(t, err)

		parsedClaims, err := ParseJWT(tokenString, secret)

		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})
}

func BenchmarkHashString(b *testing.B) {
	password := "mySecretPassword123"
	for i := 0; i < b.N; i++ {
		_, _ = HashString(password)
	}
}

func BenchmarkCheckHashString(b *testing.B) {
	password := "mySecretPassword123"
	hash, _ := HashString(password)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = CheckHashString(password, hash)
	}
}

func BenchmarkHashSHA256String(b *testing.B) {
	input := "testString"
	for i := 0; i < b.N; i++ {
		_ = HashSHA256String(input)
	}
}

func BenchmarkGenerateRandomCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomCode(10)
	}
}

func BenchmarkGenerateJWT(b *testing.B) {
	secret := []byte("test-secret-key")
	userID := uint(123)
	email := "test@example.com"
	username := "testuser"
	fullName := "Test User"

	for i := 0; i < b.N; i++ {
		_, _ = GenerateJWT(userID, secret, email, username, fullName)
	}
}
