package customerrors

// ErrorResponse is the swagger-facing shape; CustomErr's unexported HTTP fields would leak into generated specs.
type ErrorResponse struct {
	Code    string `json:"code" example:"INTERNAL_ERROR"`
	Message string `json:"message" example:"internal server error"`
}
