package handlers

import (
	"errors"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	result, err := h.userService.GetUser(c, id)
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
		Data:    result,
	})
}
