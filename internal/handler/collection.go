package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type CollectionHandler struct {
	service *service.CollectionService
}

func NewCollectionHandler() *CollectionHandler {
	return &CollectionHandler{
		service: service.NewCollectionService(),
	}
}

// List 获取集合列表
func (h *CollectionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))

	collections, total, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to get collections: "+err.Error())
		return
	}

	response.Page(c, page, perPage, total, collections)
}

// Get 通过ID获取集合
func (h *CollectionHandler) Get(c *gin.Context) {
	id := c.Param("id")

	collection, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	response.Success(c, collection)
}

// GetByName 通过名称获取集合
func (h *CollectionHandler) GetByName(c *gin.Context) {
	name := c.Param("collection")

	collection, err := h.service.GetByName(name)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	response.Success(c, collection)
}

// Create 创建集合
func (h *CollectionHandler) Create(c *gin.Context) {
	var req service.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	collection, err := h.service.Create(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, collection)
}

// Update 通过ID更新集合
func (h *CollectionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	collection, err := h.service.Update(id, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, collection)
}

// UpdateByName 通过名称更新集合
func (h *CollectionHandler) UpdateByName(c *gin.Context) {
	name := c.Param("collection")

	var req service.UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	collection, err := h.service.UpdateByName(name, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, collection)
}

// Delete 通过ID删除集合
func (h *CollectionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteByName 通过名称删除集合
func (h *CollectionHandler) DeleteByName(c *gin.Context) {
	name := c.Param("collection")

	if err := h.service.DeleteByName(name); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// CheckDelete 检查删除集合前的数据
func (h *CollectionHandler) CheckDelete(c *gin.Context) {
	name := c.Param("collection")

	result, err := h.service.CheckDelete(name)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, result)
}

// PreviewView 预览视图集合数据
func (h *CollectionHandler) PreviewView(c *gin.Context) {
	name := c.Param("collection")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))

	result, err := h.service.PreviewView(name, page, perPage)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
