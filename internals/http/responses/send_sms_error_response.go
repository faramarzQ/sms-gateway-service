package responses

type SendSMSErrorResponse struct {
	Errors []SendSMSError `json:"errors"`
}

type SendSMSError struct {
	Message  string `json:"message"`
	ClientId string `json:"client_id"`
}
