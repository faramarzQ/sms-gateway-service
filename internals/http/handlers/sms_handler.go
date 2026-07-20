package handlers

import (
	errors "errors"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type SMSHandler struct {
	smsService *services.SMSService
}

func NewSMSHandler(smsService *services.SMSService) *SMSHandler {
	return &SMSHandler{
		smsService: smsService,
	}
}

// SendSMS godoc
// @Summary      Send a single SMS
// @Description  Queues one SMS message for delivery
// @Tags         sms
// @Accept       json
// @Produce      json
// @Param        body  body      requests.SendSMSRequest  true  "SMS payload"
// @Success      200   {object}  responses.SendSMSErrorDataResponse
// @Failure      400   {object}  responses.ErrorResponse
// @Failure      402   {object}  responses.SendSMSErrorDataResponse
// @Failure      500   {object}  responses.ErrorResponse
// @Router       /sms/ [post]
func (h *SMSHandler) SendSMS(c *gin.Context) {
	var req requests.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	responseError, err := h.smsService.SendSMS(c, req)

	if err != nil {
		if httpErrors.IsDomainError(err) {
			c.JSON(http.StatusBadRequest, responses.Response{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		logger.Logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, responses.Response{
			Status:  http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    responseError,
	})
}

// SendSMSBatch godoc
// @Summary      Send SMS in batch
// @Description  Queues multiple SMS messages for delivery concurrently
// @Tags         sms
// @Accept       json
// @Produce      json
// @Param        body  body      requests.SendSMSBatchRequest  true  "Batch SMS payload"
// @Success      200   {object}  responses.SendSMSErrorDataResponse
// @Failure      400   {object}  responses.ErrorResponse
// @Failure      402   {object}  responses.SendSMSErrorDataResponse
// @Failure      500   {object}  responses.ErrorResponse
// @Router       /sms/batch [post]
func (h *SMSHandler) SendSMSBatch(c *gin.Context) {
	var req requests.SendSMSBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	responseError, err := h.smsService.SendSMSBatch(c, req)
	if err != nil {
		if errors.Is(err, httpErrors.ErrUserBalanceExceeded) {
			c.JSON(http.StatusPaymentRequired, responses.Response{
				Status:  http.StatusPaymentRequired,
				Message: err.Error(),
				Data:    responseError,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, responses.Response{
			Status:  http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    responseError,
	})
}

// GetReport godoc
// @Summary      Get SMS report for user
// @Description  Returns all SMS records for the given user
// @Tags         sms
// @Produce      json
// @Param        user_id  query     int  true  "User ID"
// @Success      200      {object}  responses.SMSReportDataResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  responses.ErrorResponse
// @Failure      500      {object}  responses.ErrorResponse
// @Router       /sms/report [get]
func (h *SMSHandler) GetReport(c *gin.Context) {
	userIdStr := c.Query("user_id")

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
		})
		return
	}

	response, err := h.smsService.GetReport(c, userId)
	if err != nil {
		switch {
		case errors.Is(err, httpErrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, responses.Response{
				Status:  http.StatusNotFound,
				Message: err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, responses.Response{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
			})
		}

		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    response,
	})

}
