package repository

import (
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"gorm.io/gorm"
)

type CollectionRepository struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) Create(collection *model.Collection) error {
	return r.db.Create(collection).Error
}

func (r *CollectionRepository) Update(collection *model.Collection) error {
	return r.db.Save(collection).Error
}

func (r *CollectionRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Collection{}, id).Error
}

func (r *CollectionRepository) GetByID(id uint64) (*model.Collection, error) {
	var collection model.Collection
	err := r.db.First(&collection, id).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (r *CollectionRepository) GetByName(name string) (*model.Collection, error) {
	var collection model.Collection
	err := r.db.First(&collection, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (r *CollectionRepository) List(page, perPage int) ([]model.Collection, int64, error) {
	var collections []model.Collection
	var total int64

	db := r.db.Model(&model.Collection{})
	db.Count(&total)

	offset := (page - 1) * perPage
	err := db.Offset(offset).Limit(perPage).Find(&collections).Error
	if err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

func (r *CollectionRepository) GetAll() ([]model.Collection, error) {
	var collections []model.Collection
	err := r.db.Find(&collections).Error
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (r *CollectionRepository) Exists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Collection{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
