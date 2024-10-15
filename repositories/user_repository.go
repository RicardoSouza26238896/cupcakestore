package repositories

import (
	"github.com/RicardoSouza26238896/cupcakestore/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindAll(filter *models.UserFilter) []models.User
	FindById(id uint) (models.User, error)
	FindByEmail(email string) (models.User, error)
	Update(user *models.User) error
	Delete(user *models.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(database *gorm.DB) UserRepository {
	return &userRepository{
		db: database,
	}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindAll(filter *models.UserFilter) []models.User {
	offset := (filter.Pagination.Page - 1) * filter.Pagination.Limit

	query := r.db.Model(&models.User{}).Omit("Password")

	if filter.User.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.User.Email+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil
	}
	filter.Pagination.Total = total

	var users []models.User
	query.Offset(offset).Limit(filter.Pagination.Limit).Order("created_at desc").Find(&users)

	return users
}

func (r *userRepository) FindById(id uint) (models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *userRepository) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *userRepository) Update(user *models.User) error {
	if user.Password == "" {
		return r.db.Omit("Password").Save(user).Error
	}
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(user *models.User) error {
	return r.db.Select("Profile").Delete(user).Error
}
