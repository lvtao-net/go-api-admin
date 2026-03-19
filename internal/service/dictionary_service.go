package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"gorm.io/gorm"
)

type DictionaryService struct {
	db *gorm.DB
}

func NewDictionaryService() *DictionaryService {
	return &DictionaryService{
		db: database.GetDB(),
	}
}

// ListResult 列表结果
type DictionaryListResult struct {
	Page       int                `json:"page"`
	PerPage    int                `json:"perPage"`
	TotalItems int64              `json:"totalItems"`
	TotalPages int                `json:"totalPages"`
	Items      []model.Dictionary `json:"items"`
}

// List 获取字典列表
func (s *DictionaryService) List(page, perPage int) (*DictionaryListResult, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 500 {
		perPage = 30
	}

	var total int64
	if err := s.db.Model(&model.Dictionary{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	var dictionaries []model.Dictionary
	if err := s.db.Order("created DESC").Offset(offset).Limit(perPage).Find(&dictionaries).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &DictionaryListResult{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      dictionaries,
	}, nil
}

// parseUintID 将字符串ID转换为uint64
func parseUintID(id string) (uint64, error) {
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", id)
	}
	return n, nil
}

// GetByID 根据ID获取字典
func (s *DictionaryService) GetByID(id string) (*model.Dictionary, error) {
	uid, err := parseUintID(id)
	if err != nil {
		return nil, err
	}
	var dict model.Dictionary
	if err := s.db.First(&dict, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dictionary not found")
		}
		return nil, err
	}
	return &dict, nil
}

// getByUintID 内部使用uint64 ID获取字典
func (s *DictionaryService) getByUintID(id uint64) (*model.Dictionary, error) {
	var dict model.Dictionary
	if err := s.db.First(&dict, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dictionary not found")
		}
		return nil, err
	}
	return &dict, nil
}

// GetByName 根据名称获取字典
func (s *DictionaryService) GetByName(name string) (*model.Dictionary, error) {
	var dict model.Dictionary
	if err := s.db.Where("name = ?", name).First(&dict).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dictionary not found")
		}
		return nil, err
	}
	return &dict, nil
}

// CreateDictionaryRequest 创建字典请求
type CreateDictionaryRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Items       []model.DictionaryItem `json:"items"`
}

// Create 创建字典
func (s *DictionaryService) Create(req *CreateDictionaryRequest) (*model.Dictionary, error) {
	// 检查名称是否已存在
	var count int64
	s.db.Model(&model.Dictionary{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("dictionary name already exists")
	}

	dict := &model.Dictionary{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		Items:       req.Items,
	}

	if dict.Label == "" {
		dict.Label = req.Name
	}

	if err := s.db.Create(dict).Error; err != nil {
		return nil, err
	}

	return dict, nil
}

// UpdateDictionaryRequest 更新字典请求
type UpdateDictionaryRequest struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Items       []model.DictionaryItem `json:"items"`
}

// Update 更新字典
func (s *DictionaryService) Update(id string, req *UpdateDictionaryRequest) (*model.Dictionary, error) {
	uid, err := parseUintID(id)
	if err != nil {
		return nil, err
	}
	dict, err := s.getByUintID(uid)
	if err != nil {
		return nil, err
	}

	// 系统字典不可修改名称
	if dict.System {
		req.Name = dict.Name
	}

	updates := map[string]interface{}{
		"label":       req.Label,
		"description": req.Description,
		"items":       req.Items,
	}

	if err := s.db.Model(dict).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.getByUintID(uid)
}

// Delete 删除字典
func (s *DictionaryService) Delete(id string) error {
	uid, err := parseUintID(id)
	if err != nil {
		return err
	}
	dict, err := s.getByUintID(uid)
	if err != nil {
		return err
	}

	if dict.System {
		return errors.New("system dictionary cannot be deleted")
	}

	return s.db.Delete(dict).Error
}

// GetItems 获取字典项列表
func (s *DictionaryService) GetItems(id string) ([]model.DictionaryItem, error) {
	dict, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	return dict.Items, nil
}

// CreateDictionaryItemRequest 创建字典项请求
type CreateDictionaryItemRequest struct {
	Label       string `json:"label" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Sort        int    `json:"sort"`
	Disabled    bool   `json:"disabled"`
	Description string `json:"description"`
}

// CreateItem 创建字典项
func (s *DictionaryService) CreateItem(id string, req *CreateDictionaryItemRequest) (*model.DictionaryItem, error) {
	dict, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 生成自增ID（取当前最大ID + 1）
	var maxID uint64
	for _, item := range dict.Items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}

	item := model.DictionaryItem{
		ID:          maxID + 1,
		Label:       req.Label,
		Value:       req.Value,
		Sort:        req.Sort,
		Disabled:    req.Disabled,
		Description: req.Description,
	}

	dict.Items = append(dict.Items, item)

	if err := s.db.Model(dict).Update("items", dict.Items).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// UpdateDictionaryItemRequest 更新字典项请求
type UpdateDictionaryItemRequest struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Sort        int    `json:"sort"`
	Disabled    bool   `json:"disabled"`
	Description string `json:"description"`
}

// UpdateItem 更新字典项
func (s *DictionaryService) UpdateItem(id string, itemID string, req *UpdateDictionaryItemRequest) (*model.DictionaryItem, error) {
	dict, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	iid, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return nil, err
	}

	var targetItem *model.DictionaryItem
	for i := range dict.Items {
		if dict.Items[i].ID == iid {
			targetItem = &dict.Items[i]
			break
		}
	}

	if targetItem == nil {
		return nil, errors.New("item not found")
	}

	if req.Label != "" {
		targetItem.Label = req.Label
	}
	if req.Value != "" {
		targetItem.Value = req.Value
	}
	targetItem.Sort = req.Sort
	targetItem.Disabled = req.Disabled
	if req.Description != "" {
		targetItem.Description = req.Description
	}

	if err := s.db.Model(dict).Update("items", dict.Items).Error; err != nil {
		return nil, err
	}

	return targetItem, nil
}

// DeleteItem 删除字典项
func (s *DictionaryService) DeleteItem(id string, itemID string) error {
	dict, err := s.GetByID(id)
	if err != nil {
		return err
	}

	iid, err := parseUintID(itemID)
	if err != nil {
		return err
	}

	found := false
	newItems := make([]model.DictionaryItem, 0)
	for _, item := range dict.Items {
		if item.ID == iid {
			found = true
			continue
		}
		newItems = append(newItems, item)
	}

	if !found {
		return errors.New("item not found")
	}

	dict.Items = newItems
	return s.db.Model(dict).Update("items", dict.Items).Error
}

// InitSystemDictionaries 初始化系统字典
func (s *DictionaryService) InitSystemDictionaries() error {
	systemDicts := []model.Dictionary{
		{
			Name:        "field_type",
			Label:       "字段类型",
			Description: "系统字段类型",
			System:      true,
			Items: []model.DictionaryItem{
				{ID: 1, Label: "文本", Value: "text", Sort: 1},
				{ID: 2, Label: "数字", Value: "number", Sort: 2},
				{ID: 3, Label: "布尔", Value: "bool", Sort: 3},
				{ID: 4, Label: "邮箱", Value: "email", Sort: 4},
				{ID: 5, Label: "URL", Value: "url", Sort: 5},
				{ID: 6, Label: "日期时间", Value: "date", Sort: 6},
				{ID: 7, Label: "单选", Value: "radio", Sort: 7},
				{ID: 8, Label: "多选", Value: "checkbox", Sort: 8},
				{ID: 9, Label: "下拉选择", Value: "select", Sort: 9},
				{ID: 10, Label: "关联关系", Value: "relation", Sort: 10},
				{ID: 11, Label: "文件", Value: "file", Sort: 11},
				{ID: 12, Label: "富文本", Value: "editor", Sort: 12},
				{ID: 13, Label: "JSON", Value: "json", Sort: 13},
				{ID: 14, Label: "密码", Value: "password", Sort: 14},
			},
		},
	}

	for _, dict := range systemDicts {
		var existing model.Dictionary
		err := s.db.Where("name = ?", dict.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.Create(&dict).Error; err != nil {
				return fmt.Errorf("failed to create system dictionary %s: %w", dict.Name, err)
			}
		}
	}

	return nil
}
