package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalItems int64       `json:"totalItems"`
	TotalPages int         `json:"totalPages"`
	Items      interface{} `json:"items"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	httpCode := http.StatusBadRequest
	if code >= 500 {
		httpCode = http.StatusInternalServerError
	} else if code == 401 {
		httpCode = http.StatusUnauthorized
	} else if code == 403 {
		httpCode = http.StatusForbidden
	} else if code == 404 {
		httpCode = http.StatusNotFound
	}

	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	httpCode := http.StatusBadRequest
	if code >= 500 {
		httpCode = http.StatusInternalServerError
	} else if code == 401 {
		httpCode = http.StatusUnauthorized
	} else if code == 403 {
		httpCode = http.StatusForbidden
	} else if code == 404 {
		httpCode = http.StatusNotFound
	}

	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, 403, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}

func Page(c *gin.Context, page, perPage int, total int64, items interface{}) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	Success(c, PageData{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      items,
	})
}
