package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/internal/transaction"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type RecordHandler struct {
	collectionService *service.CollectionService
	recordService     *service.RecordService
}

func NewRecordHandler() *RecordHandler {
	return &RecordHandler{
		collectionService: service.NewCollectionService(),
		recordService:     service.NewRecordService(),
	}
}

// isAdmin 检查是否是管理员
func isAdmin(c *gin.Context) bool {
	_, hasAdminID := c.Get("admin_id")
	return hasAdminID
}

// List 获取记录列表
func (h *RecordHandler) List(c *gin.Context) {
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

	// 根据字段权限过滤数据
	if !isAdmin(c) {
		for i, item := range result.Items {
			result.Items[i].Data = service.FilterFieldsForList(item.Data, collection.Fields, false)
		}
	}

	response.Success(c, result)
}

// Get 获取单条记录
func (h *RecordHandler) Get(c *gin.Context) {
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

	// 根据字段权限过滤数据
	if !isAdmin(c) {
		record.Data = service.FilterFieldsForView(record.Data, collection.Fields, false)
	}

	response.Success(c, record)
}

// GetByField 通过指定字段获取单条记录
func (h *RecordHandler) GetByField(c *gin.Context) {
	collectionName := c.Param("collection")
	field := c.Param("field")
	value := c.Param("value")

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

	// 检查字段是否允许查找
	allowedFields := h.getAllowedLookupFields(collection)
	isAllowed := false
	for _, f := range allowedFields {
		if f == field {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		response.BadRequest(c, "Field '"+field+"' is not allowed for lookup")
		return
	}

	// 通过字段查找记录
	record, err := h.recordService.GetByField(collection, field, value)
	if err != nil {
		response.NotFound(c, "Record not found")
		return
	}

	// 根据字段权限过滤数据
	if !isAdmin(c) {
		record.Data = service.FilterFieldsForView(record.Data, collection.Fields, false)
	}

	response.Success(c, record)
}

// getAllowedLookupFields 获取允许查找的字段列表
func (h *RecordHandler) getAllowedLookupFields(collection *model.Collection) []string {
	fields := []string{}

	// 主键字段始终允许
	pkField := collection.PrimaryKeyField
	if pkField == "" {
		pkField = "id"
	}
	fields = append(fields, pkField)

	// 添加配置的可查找字段
	for _, lf := range collection.LookupFields {
		fields = append(fields, lf.Field)
	}

	// 添加唯一字段
	for _, f := range collection.Fields {
		if f.Unique && f.Name != pkField {
			fields = append(fields, f.Name)
		}
	}

	return fields
}

// Create 创建记录
func (h *RecordHandler) Create(c *gin.Context) {
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

	// 根据字段权限过滤输入数据
	if !isAdmin(c) {
		data = service.FilterFieldsForCreate(data, collection.Fields, false)
	}

	// 创建记录
	record, err := h.recordService.Create(collection, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 根据字段权限过滤输出数据
	if !isAdmin(c) {
		record.Data = service.FilterFieldsForView(record.Data, collection.Fields, false)
	}

	response.Success(c, record)
}

// Update 更新记录
func (h *RecordHandler) Update(c *gin.Context) {
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

	// 获取当前用户信息
	userID, hasUserID := c.Get("user_id")
	_, hasAdminID := c.Get("admin_id")

	// 检查权限：管理员可操作所有，用户只能操作自己的记录
	if hasUserID && !hasAdminID && collection.Type == "auth" {
		// Auth类型的集合，用户只能修改自己的记录
		var uid uint
		switch v := userID.(type) {
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		case string:
			// JWT 中存储的是字符串
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				response.Forbidden(c, "Invalid user ID")
				return
			}
			uid = uint(parsed)
		default:
			response.Forbidden(c, "Invalid user ID type")
			return
		}
		if uid != uint(id) {
			response.Forbidden(c, "You can only update your own record")
			return
		}
	}

	// 解析请求体
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 根据字段权限过滤输入数据
	if !isAdmin(c) {
		data = service.FilterFieldsForUpdate(data, collection.Fields, false)
	}

	// 更新记录
	record, err := h.recordService.Update(collection, id, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 根据字段权限过滤输出数据
	if !isAdmin(c) {
		record.Data = service.FilterFieldsForView(record.Data, collection.Fields, false)
	}

	response.Success(c, record)
}

// Delete 删除记录
func (h *RecordHandler) Delete(c *gin.Context) {
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

	// 获取当前用户信息
	userID, hasUserID := c.Get("user_id")
	_, hasAdminID := c.Get("admin_id")

	// 检查权限：管理员可操作所有，用户只能操作自己的记录
	if hasUserID && !hasAdminID && collection.Type == "auth" {
		// Auth类型的集合，用户只能删除自己的记录
		var uid uint
		switch v := userID.(type) {
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		}
		if uid != uint(id) {
			response.Forbidden(c, "You can only delete your own record")
			return
		}
	}

	// 删除记录
	if err := h.recordService.Delete(collection, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchDelete 批量删除记录
func (h *RecordHandler) BatchDelete(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
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

// parsePage 解析分页参数
func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 500 {
		perPage = 30
	}

	return page, perPage
}

// GetCollectionFields 获取集合字段（供前端使用）
func (h *RecordHandler) GetCollectionFields(c *gin.Context) {
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

	// 过滤字段，只返回 API 可访问的字段
	fields := make([]model.CollectionField, 0)
	for _, field := range collection.Fields {
		if field.APIDisabled {
			continue
		}
		fields = append(fields, field)
	}

	response.Success(c, gin.H{
		"name":   collection.Name,
		"label":  collection.Label,
		"type":   collection.Type,
		"fields": fields,
	})
}

// ViewQuery 执行视图查询
func (h *RecordHandler) ViewQuery(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 只有 view 类型才能执行查询
	if collection.Type != "view" {
		response.BadRequest(c, "Only view collections support query operation")
		return
	}

	// 解析查询参数
	var req service.ListRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// 执行视图查询
	result, err := h.recordService.ViewQuery(collection, &req)
	if err != nil {
		response.InternalError(c, "Failed to execute view query: "+err.Error())
		return
	}

	response.Success(c, result)
}

// TransactionExecute 执行事务
func (h *RecordHandler) TransactionExecute(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 只有 transaction 类型才能执行
	if collection.Type != "transaction" {
		response.BadRequest(c, "Only transaction collections support execute operation")
		return
	}

	// 解析请求体
	var req struct {
		Params map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 获取当前用户ID
	userID, _ := c.Get("user_id")
	var uid uint
	if uidVal, ok := userID.(uint); ok {
		uid = uidVal
	} else if uidVal, ok := userID.(float64); ok {
		uid = uint(uidVal)
	}

	// 执行事务
	transactionService := transaction.NewService()
	result, err := transactionService.Execute(collectionName, req.Params, uid)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, result)
}
