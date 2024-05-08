package repositories

import (
	"go_final/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Getproduct(int) (models.Product, error)
	GetAllproduct() ([]models.Product, error)
	AddProduct(models.Product) (models.Product, error)
	UpdateProduct(models.Product) (models.Product, error)
	DeleteProduct(models.Product) (models.Product, error)
}

type productRepository struct {
	connection *gorm.DB
}

func NewProductRepository() ProductRepository {
	return &productRepository{
		connection: DB(),
	}
}

func (db *productRepository) Getproduct(id int) (product models.Product, err error) {
	return product, db.connection.First(&product, id).Error
}

func (db *productRepository) GetAllproduct() (products []models.Product, err error) {
	return products, db.connection.Find(&products).Error
}

func (db *productRepository) AddProduct(product models.Product) (models.Product, error) {
	return product, db.connection.Create(&product).Error
}

func (db *productRepository) UpdateProduct(product models.Product) (models.Product, error) {
	if err := db.connection.First(&models.Product{}, product.ID).Error; err != nil {
		return product, err
	}
	return product, db.connection.Model(&product).Updates(&product).Error
}

func (db *productRepository) DeleteProduct(product models.Product) (models.Product, error) {
	if err := db.connection.First(&product, product.ID).Error; err != nil {
		return product, err
	}
	return product, db.connection.Unscoped().Delete(&product).Error
}
