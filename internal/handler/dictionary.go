package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type DictionaryHandler struct {
	service *service.DictionaryService
}

func NewDictionaryHandler() *DictionaryHandler {
	return &DictionaryHandler{
		service: service.NewDictionaryService(),
	}
}

// List 获取字典列表
func (h *DictionaryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))

	result, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to list dictionaries: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Get 获取单个字典
func (h *DictionaryHandler) Get(c *gin.Context) {
	id := c.Param("id")

	dict, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "Dictionary not found")
		return
	}

	response.Success(c, dict)
}

// GetByName 根据名称获取字典
func (h *DictionaryHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	dict, err := h.service.GetByName(name)
	if err != nil {
		response.NotFound(c, "Dictionary not found")
		return
	}

	response.Success(c, dict)
}

// Create 创建字典
func (h *DictionaryHandler) Create(c *gin.Context) {
	var req service.CreateDictionaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	dict, err := h.service.Create(&req)
	if err != nil {
		response.InternalError(c, "Failed to create dictionary: "+err.Error())
		return
	}

	response.Success(c, dict)
}

// Update 更新字典
func (h *DictionaryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateDictionaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	dict, err := h.service.Update(id, &req)
	if err != nil {
		response.InternalError(c, "Failed to update dictionary: "+err.Error())
		return
	}

	response.Success(c, dict)
}

// Delete 删除字典
func (h *DictionaryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(id); err != nil {
		response.InternalError(c, "Failed to delete dictionary: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetItems 获取字典项列表
func (h *DictionaryHandler) GetItems(c *gin.Context) {
	id := c.Param("id")

	items, err := h.service.GetItems(id)
	if err != nil {
		response.InternalError(c, "Failed to get dictionary items: "+err.Error())
		return
	}

	response.Success(c, items)
}

// CreateItem 创建字典项
func (h *DictionaryHandler) CreateItem(c *gin.Context) {
	id := c.Param("id")

	var req service.CreateDictionaryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.CreateItem(id, &req)
	if err != nil {
		response.InternalError(c, "Failed to create dictionary item: "+err.Error())
		return
	}

	response.Success(c, item)
}

// UpdateItem 更新字典项
func (h *DictionaryHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	var req service.UpdateDictionaryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	item, err := h.service.UpdateItem(id, itemID, &req)
	if err != nil {
		response.InternalError(c, "Failed to update dictionary item: "+err.Error())
		return
	}

	response.Success(c, item)
}

// DeleteItem 删除字典项
func (h *DictionaryHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	if err := h.service.DeleteItem(id, itemID); err != nil {
		response.InternalError(c, "Failed to delete dictionary item: "+err.Error())
		return
	}

	response.Success(c, nil)
}
