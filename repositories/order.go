package repositories

import (
	"errors"
	"fmt"
	"go_final/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	GetOrders(uint) ([]models.Order, error)
	GetOrderByID(uint) (models.Order, error)
	GetOrderItems(uint) ([]models.OrderItems, error)
	OrderProducts(uint, []models.CartItemRequest) error
	UpdateOrder(uint, []models.CartItemRequest) error
	UpdateOrderStatus(uint, string) error
	DeleteOrder(uint) error
	DeleteOrderItem(uint, uint) error
}

type orderRepository struct {
	connection *gorm.DB
}

func NewOrderRepository() OrderRepository {
	return &orderRepository{
		connection: DB(),
	}
}

func (db *orderRepository) GetOrders(userID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := db.connection.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (db *orderRepository) GetOrderByID(orderID uint) (models.Order, error) {
	var order models.Order
	if err := db.connection.First(&order, orderID).Error; err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (db *orderRepository) GetOrderItems(orderID uint) ([]models.OrderItems, error) {
	var orderItems []models.OrderItems
	err := db.connection.Preload(clause.Associations).Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.id = ?", orderID).
		Find(&orderItems).Error
	if err != nil {
		return nil, err
	}
	return orderItems, nil
}

func addNewOrderItem(tx *gorm.DB, orderID uint, item models.CartItemRequest) error {
	var product models.Product
	if err := tx.First(&product, item.ProductID).Error; err != nil {
		return err
	}

	if product.Quantity < item.Quantity {
		return errors.New("not enough stock available for: " + product.Name)
	}

	newQuantity := product.Quantity - item.Quantity
	if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).Update("quantity", newQuantity).Error; err != nil {
		return err
	}

	orderItem := models.OrderItems{
		OrderID:   orderID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		Price:     product.Price,
	}

	if err := tx.Create(&orderItem).Error; err != nil {
		return err
	}

	return nil
}

func (db *orderRepository) OrderProducts(userID uint, request []models.CartItemRequest) error {
	return db.connection.Transaction(func(tx *gorm.DB) error {
		var newOrder models.Order
		result := tx.Where("user_id = ? AND (order_status = ? OR order_status = ?)", userID, models.PENDING, models.ACCEPTED).FirstOrCreate(&newOrder, models.Order{UserID: userID, OrderStatus: models.PENDING})
		if result.Error != nil {
			return result.Error
		}

		for _, item := range request {
			if err := addNewOrderItem(tx, newOrder.ID, item); err != nil {
				return err
			}
		}

		return nil
	})
}

func (db *orderRepository) UpdateOrder(userID uint, request []models.CartItemRequest) error {
	return db.connection.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		result := tx.Where("user_id = ? AND (order_status = ? OR order_status = ?)", userID, models.PENDING, models.ACCEPTED).First(&order)
		if result.Error != nil {
			return result.Error
		}

		// заказ до обновления
		var existingItems []models.OrderItems
		if err := tx.Where("order_id = ?", order.ID).Find(&existingItems).Error; err != nil {
			return err
		}

		// мапим по айдишкам чтобы потом легче обращаться
		existingMap := make(map[uint]models.OrderItems)
		for _, item := range existingItems {
			existingMap[item.ID] = item
		}

		// пробегаемся по каждому элементу апдейта заказа
		for _, newItem := range request {
			existingItem, exists := existingMap[newItem.OrderItemID]

			// обновляем существующий элемент в заказе (т.е. этот продукт был до обновления
			// или нужно добавить новый продукт в заказ)
			if exists {
				// если этот продукт нужно полностью заменить другим
				if existingItem.ProductID != newItem.ProductID {
					// проверяем наличие для нового продукта и уменьшаем количество в базе
					if err := checkAndAdjustStock(tx, newItem.ProductID, -newItem.Quantity); err != nil {
						return err
					}
					// возвращаем на склад старый товар и увеличиваем количество в базе без проверки
					if err := adjustInventory(tx, existingItem.ProductID, existingItem.Quantity); err != nil {
						return err
					}
					// если это тот же товар и юзер просто меняет его количество
				} else if existingItem.Quantity != newItem.Quantity {
					// меняем количество на складе
					difference := newItem.Quantity - existingItem.Quantity
					// -diff потому что когда юзер уменьшил количество товара мы должны вернуть эту разницу на склад
					if err := checkAndAdjustStock(tx, newItem.ProductID, -difference); err != nil {
						return err
					}
				}

				// сохраняем изменения
				existingItem.ProductID = newItem.ProductID // на случай если товар поменялся
				existingItem.Quantity = newItem.Quantity
				if err := tx.Save(&existingItem).Error; err != nil {
					return err
				}
			} else {
				// просто доваляем новый товар
				if err := addNewOrderItem(tx, order.ID, newItem); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (db *orderRepository) UpdateOrderStatus(orderID uint, newStatus string) error {
	return db.connection.Model(&models.Order{}).Where("id = ?", orderID).Update("order_status", newStatus).Error
}

func checkAndAdjustStock(tx *gorm.DB, productID uint, quantityChange int) error {
	var product models.Product
	if err := tx.First(&product, productID).Error; err != nil {
		return err
	}

	newQuantity := product.Quantity + quantityChange
	if newQuantity < 0 {
		return errors.New("not enough stock available for product ID " + fmt.Sprint(productID))
	}

	return tx.Model(&models.Product{}).Where("id = ?", productID).Update("quantity", newQuantity).Error
}

func adjustInventory(tx *gorm.DB, productID uint, quantity int) error {
	return tx.Model(&models.Product{}).Where("id = ?", productID).Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
}

func (db *orderRepository) DeleteOrder(userID uint) error {
	return db.connection.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Where("user_id = ? AND (order_status = ? OR order_status = ?)", userID, models.PENDING, models.ACCEPTED).First(&order).Error; err != nil {
			return err
		}

		var orderItems []models.OrderItems
		if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
			return err
		}

		for _, item := range orderItems {
			if err := adjustInventory(tx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}

		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItems{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&order).Error; err != nil {
			return err
		}

		return nil
	})
}

func (db *orderRepository) DeleteOrderItem(userID uint, orderItemID uint) error {
	return db.connection.Transaction(func(tx *gorm.DB) error {
		var orderItem models.OrderItems
		result := tx.Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("order_items.id = ? AND orders.user_id = ?", orderItemID, userID).
			First(&orderItem)
		if result.Error != nil {
			return result.Error
		}

		if err := tx.Delete(&orderItem).Error; err != nil {
			return err
		}

		if err := adjustInventory(tx, orderItem.ProductID, orderItem.Quantity); err != nil {
			return err
		}

		return nil
	})
}
