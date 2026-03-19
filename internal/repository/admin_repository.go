package repository

import (
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepository) Update(admin *model.Admin) error {
	return r.db.Save(admin).Error
}

func (r *AdminRepository) Delete(id string) error {
	return r.db.Delete(&model.Admin{}, "id = ?", id).Error
}

func (r *AdminRepository) GetByID(id string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) GetByEmail(email string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.First(&admin, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) List(page, perPage int) ([]model.Admin, int64, error) {
	var admins []model.Admin
	var total int64

	db := r.db.Model(&model.Admin{})
	db.Count(&total)

	offset := (page - 1) * perPage
	err := db.Offset(offset).Limit(perPage).Find(&admins).Error
	if err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

func (r *AdminRepository) Exists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *AdminRepository) UpdateTokenKey(id, tokenKey string) error {
	return r.db.Model(&model.Admin{}).Where("id = ?", id).Update("token_key", tokenKey).Error
}
