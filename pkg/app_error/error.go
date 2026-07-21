package apperror

type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    Details     string `json:"details,omitempty"`
    StatusCode int    `json:"status_code"`
    ErrorData   any `json:"error_data,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

func New(status int, code, message string, details string, errorData any) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        Details:    details,
        StatusCode: status,
        ErrorData:  errorData,
    }
}
