package handlers

import (
	"errors"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
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
		c.JSON(http.StatusInternalServerError, responses.Response{
			Status:  http.StatusInternalServerError,
			Message: "internal server error",
		})
	}

	c.JSON(http.StatusOK, responses.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    responseError,
	})
}

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
		c.JSON(http.StatusInternalServerError, responses.Response{
			Status:  http.StatusInternalServerError,
			Message: "internal server error",
		})
	}

	c.JSON(http.StatusOK, responses.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    responseError,
	})
}

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
