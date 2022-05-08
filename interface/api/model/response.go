package model

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	ErrCode int    `json:"errCode"`
	Detail  string `json:"detail"`
}
