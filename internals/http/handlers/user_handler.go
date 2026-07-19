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

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Returns user details including balance and traffic class
// @Tags         user
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  responses.UserDataResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /user/{id} [get]
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

// IncreaseBalance godoc
// @Summary      Increase user balance
// @Description  Adds credits to a user's SMS balance
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id    path      int                              true  "User ID"
// @Param        body  body      requests.IncreaseBalanceRequest  true  "Amount to add"
// @Success      200   {object}  responses.EmptyDataResponse
// @Failure      400   {object}  responses.ErrorResponse
// @Failure      404   {object}  responses.ErrorResponse
// @Failure      500   {object}  responses.ErrorResponse
// @Router       /user/{id}/balance [put]
func (h *UserHandler) IncreaseBalance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	var req requests.IncreaseBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, responses.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = h.userService.IncreaseBalance(c, id, req)
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
	})
}
