package repositories

import (
	"go_final/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(models.User) (models.User, error)
	GetUser(int) (models.APIUser, error)
	GetByEmail(string) (models.User, error)
	GetAllUsers() ([]models.APIUser, error)
	UpdateUser(models.User) (models.User, error)
	DeleteUser(models.User) (models.User, error)
}

type userRepository struct {
	connection *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{
		connection: DB(),
	}
}

func (db *userRepository) GetUser(id int) (user models.APIUser, err error) {
	return user, db.connection.Model(&models.User{}).First(&user, id).Error
}

func (db *userRepository) GetByEmail(email string) (user models.User, err error) {
	return user, db.connection.First(&user, "email=?", email).Error
}

func (db *userRepository) GetAllUsers() (users []models.APIUser, err error) {
	return users, db.connection.Model(&models.User{}).Find(&users).Error
}

func (db *userRepository) CreateUser(user models.User) (models.User, error) {
	return user, db.connection.Create(&user).Error
}

func (db *userRepository) UpdateUser(user models.User) (models.User, error) {
	if err := db.connection.First(&models.User{}, user.ID).Error; err != nil {
		return user, err
	}
	return user, db.connection.Model(&user).Updates(&user).Error
}

func (db *userRepository) DeleteUser(user models.User) (models.User, error) {
	if err := db.connection.First(&user, user.ID).Error; err != nil {
		return user, err
	}
	return user, db.connection.Unscoped().Delete(&user).Error
}
