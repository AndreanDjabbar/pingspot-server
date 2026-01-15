package env

import "os"

func Port() string            { return os.Getenv("PORT") }
func Host() string            { return os.Getenv("HOST") }
func LogLevel() string        { return os.Getenv("LOG_LEVEL") }
func NodeEnv() string         { return os.Getenv("NODE_ENV") }
func ClientURL() string       { return os.Getenv("CLIENT_URL") }
func GithubRepoURL() string   { return os.Getenv("GITHUB_REPO_URL") }
func PostgreHost() string     { return os.Getenv("POSTGRE_HOST") }
func PostgrePort() string     { return os.Getenv("POSTGRE_PORT") }
func PostgreUser() string     { return os.Getenv("POSTGRE_USER") }
func PostgrePassword() string { return os.Getenv("POSTGRE_PASSWORD") }
func PostgreDB() string       { return os.Getenv("POSTGRE_DB") }
func RedisHost() string       { return os.Getenv("REDIS_HOST") }
func RedisPort() string       { return os.Getenv("REDIS_PORT") }
func JWTSecret() string       { return os.Getenv("JWT_SECRET") }
func EmailPassword() string   { return os.Getenv("EMAIL_PASSWORD") }
func EmailEmail() string      { return os.Getenv("EMAIL_EMAIL") }
func GoogleClientID() string { return os.Getenv("GOOGLE_CLIENT_ID") }
func GoogleClientSecret() string { return os.Getenv("GOOGLE_CLIENT_SECRET") }
func GoogleCallbackURL() string { return os.Getenv("GOOGLE_CALLBACK_URL") }
func GoogleSecretSessionKey() string { return os.Getenv("GOOGLE_SECRET_SESSION_KEY") }
func IsProduction() bool { return os.Getenv("IS_PRODUCTION") == "true" }
func IsHTTPOnly() bool { return os.Getenv("HTTP_ONLY") == "true" }
func MongoUser() string { return os.Getenv("MONGO_USER") }
func MongoHost() string { return os.Getenv("MONGO_HOST") }
func MongoPort() string { return os.Getenv("MONGO_PORT") }
func MongoPassword() string { return os.Getenv("MONGO_PASSWORD") }
func AccessTokenAge() string { return os.Getenv("ACCESS_TOKEN_AGE") }
func RefreshTokenAge() string { return os.Getenv("REFRESH_TOKEN_AGE") }
func GlobalRateLimiterMaxRequests() string {
	return os.Getenv("GLOBAL_RATE_LIMITER_MAX_REQUESTS")
}
func GlobalRateLimiterWindowSeconds() string {
	return os.Getenv("GLOBAL_RATE_LIMITER_WINDOW_SECONDS")
}
func RedisUsername() string { return os.Getenv("REDIS_USERNAME") }
func RedisPassword() string { return os.Getenv("REDIS_PASSWORD") }