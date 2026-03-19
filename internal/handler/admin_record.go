package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/password"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

// AdminRecordHandler 后台管理记录处理器（不受 API 权限限制）
type AdminRecordHandler struct {
	collectionService *service.CollectionService
	recordService     *service.RecordService
}

// NewAdminRecordHandler 创建后台记录处理器
func NewAdminRecordHandler() *AdminRecordHandler {
	return &AdminRecordHandler{
		collectionService: service.NewCollectionService(),
		recordService:     service.NewRecordService(),
	}
}

// List 获取记录列表（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) List(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表，不提供记录管理
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 解析请求参数
	var req service.ListRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// 获取记录列表
	result, err := h.recordService.List(collection, &req)
	if err != nil {
		response.InternalError(c, "Failed to list records: "+err.Error())
		return
	}

	// 后台管理：隐藏密码字段内容
	h.hidePasswordFields(collection, result)

	response.Success(c, result)
}

// Get 获取单条记录（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) Get(c *gin.Context) {
	collectionName := c.Param("collection")
	idStr := c.Param("id")

	// 解析ID
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid record ID")
		return
	}

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 获取记录
	record, err := h.recordService.GetByID(collection, id)
	if err != nil {
		response.NotFound(c, "Record not found")
		return
	}

	// 后台管理：隐藏密码字段内容
	h.hidePasswordFieldsInRecord(collection, record)

	response.Success(c, record)
}

// Create 创建记录（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) Create(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 解析请求体
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 后台管理：处理密码字段加密
	h.hashPasswordFields(collection, data)

	// 创建记录
	record, err := h.recordService.Create(collection, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 后台管理：隐藏密码字段内容
	h.hidePasswordFieldsInRecord(collection, record)

	response.Success(c, record)
}

// Update 更新记录（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) Update(c *gin.Context) {
	collectionName := c.Param("collection")
	idStr := c.Param("id")

	// 解析ID
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid record ID")
		return
	}

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 解析请求体
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 后台管理：处理密码字段（非空加密，空则删除该字段不更新）
	h.processPasswordFieldsForUpdate(collection, data)

	// 更新记录
	record, err := h.recordService.Update(collection, id, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 后台管理：隐藏密码字段内容
	h.hidePasswordFieldsInRecord(collection, record)

	response.Success(c, record)
}

// Delete 删除记录（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) Delete(c *gin.Context) {
	collectionName := c.Param("collection")
	idStr := c.Param("id")

	// 解析ID
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid record ID")
		return
	}

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 后台管理允许删除任何记录

	// 删除记录
	if err := h.recordService.Delete(collection, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchDelete 批量删除记录（后台管理，不受 API 权限限制）
func (h *AdminRecordHandler) BatchDelete(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际表
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not support record management")
		return
	}

	// 解析请求体
	var req struct {
		Ids []uint64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 批量删除
	if err := h.recordService.BatchDelete(collection, req.Ids); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetCollectionFields 获取集合字段（后台管理，显示所有字段）
func (h *AdminRecordHandler) GetCollectionFields(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 视图和事务集合没有实际字段
	if collection.Type == "view" || collection.Type == "transaction" {
		response.BadRequest(c, "View and Transaction collections do not have fields")
		return
	}

	// 后台管理显示所有字段，不过滤
	response.Success(c, gin.H{
		"name":   collection.Name,
		"label":  collection.Label,
		"type":   collection.Type,
		"fields": collection.Fields,
	})
}

// getPasswordFieldNames 获取集合中所有密码类型的字段名
func (h *AdminRecordHandler) getPasswordFieldNames(collection *model.Collection) []string {
	var passwordFields []string
	for _, field := range collection.Fields {
		if field.Type == "password" {
			passwordFields = append(passwordFields, field.Name)
		}
	}
	return passwordFields
}

// hidePasswordFields 清空列表结果中的密码字段内容（返回空字符串，避免前端误提交）
func (h *AdminRecordHandler) hidePasswordFields(collection *model.Collection, result *service.ListResult) {
	passwordFields := h.getPasswordFieldNames(collection)
	if len(passwordFields) == 0 {
		return
	}

	for _, item := range result.Items {
		for _, fieldName := range passwordFields {
			if _, exists := item.Data[fieldName]; exists {
				item.Data[fieldName] = "" // 返回空字符串，用户不填则不修改
			}
		}
	}
}

// hidePasswordFieldsInRecord 清空单条记录中的密码字段内容（返回空字符串，避免前端误提交）
func (h *AdminRecordHandler) hidePasswordFieldsInRecord(collection *model.Collection, record *service.RecordResult) {
	if record == nil || record.Data == nil {
		return
	}

	passwordFields := h.getPasswordFieldNames(collection)
	for _, fieldName := range passwordFields {
		if _, exists := record.Data[fieldName]; exists {
			record.Data[fieldName] = "" // 返回空字符串，用户不填则不修改
		}
	}
}

// hashPasswordFields 对密码字段进行加密（用于创建）
func (h *AdminRecordHandler) hashPasswordFields(collection *model.Collection, data map[string]interface{}) {
	passwordFields := h.getPasswordFieldNames(collection)
	for _, fieldName := range passwordFields {
		if value, exists := data[fieldName]; exists {
			if strVal, ok := value.(string); ok && strVal != "" {
				hashedPassword, err := password.HashPassword(strVal)
				if err == nil {
					data[fieldName] = hashedPassword
				}
			}
		}
	}
}

// processPasswordFieldsForUpdate 处理更新时的密码字段（非空加密，空则删除不更新）
func (h *AdminRecordHandler) processPasswordFieldsForUpdate(collection *model.Collection, data map[string]interface{}) {
	passwordFields := h.getPasswordFieldNames(collection)
	for _, fieldName := range passwordFields {
		if value, exists := data[fieldName]; exists {
			if strVal, ok := value.(string); ok {
				if strVal == "" {
					delete(data, fieldName)
				} else {
					hashedPassword, err := password.HashPassword(strVal)
					if err == nil {
						data[fieldName] = hashedPassword
					}
				}
			}
		}
	}
}
