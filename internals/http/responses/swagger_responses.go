package responses

import "github.com/faramarzQ/sms-gateway-service/internals/models"

type UserDataResponse struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"ok"`
	Data    models.User `json:"data"`
}

type SMSReportDataResponse struct {
	Status  int                   `json:"status" example:"200"`
	Message string                `json:"message" example:"ok"`
	Data    UserSMSReportResponse `json:"data"`
}

type SendSMSErrorDataResponse struct {
	Status  int                  `json:"status" example:"200"`
	Message string               `json:"message" example:"ok"`
	Data    SendSMSErrorResponse `json:"data"`
}

type EmptyDataResponse struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"ok"`
}

type ErrorResponse struct {
	Status  int    `json:"status" example:"500"`
	Message string `json:"message" example:"internal server error"`
}
