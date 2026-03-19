package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/transaction"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

// TransactionHandler 事务处理器
type TransactionHandler struct {
	service *transaction.Service
}

// NewTransactionHandler 创建事务处理器
func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{
		service: transaction.NewService(),
	}
}

// ExecuteTransaction 执行事务
// @Summary 执行事务
// @Description 执行指定集合的事务操作，支持原子性多步骤操作
// @Tags 事务
// @Accept json
// @Produce json
// @Param collection path string true "事务集合名称"
// @Param request body TransactionExecuteRequest true "事务请求"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/collections/{collection}/transactions/execute [post]
func (h *TransactionHandler) ExecuteTransaction(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未授权")
		return
	}

	collectionName := c.Param("collection")
	if collectionName == "" {
		response.BadRequest(c, "集合名称不能为空")
		return
	}

	var req TransactionExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	result, err := h.service.Execute(collectionName, req.Params, userID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, result)
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) uint {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	switch v := userIDInterface.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	case int:
		return uint(v)
	case string:
		// 尝试解析字符串
		var parsed uint
		if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}

// TransactionExecuteRequest 事务执行请求
type TransactionExecuteRequest struct {
	Params map[string]interface{} `json:"params"`
}
