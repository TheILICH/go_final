package route

import (
	"github.com/gin-gonic/gin"
	"go_final/handlers"
	"go_final/middleware"
	"net/http"
)

func RunAPI(address string) error {

	userHandler := handlers.NewUserHandler()
	productHandler := handlers.NewProductHandler()
	orderHandler := handlers.NewOrderHandler()

	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Welcome to SDU Canteen!")
	})

	apiRoutes := r.Group("/api")
	userRoutes := apiRoutes.Group("/user")
	{
		userRoutes.POST("/register", userHandler.CreateUser)
		userRoutes.POST("/signin", userHandler.SignInUser)
	}

	userSecuredRoutes := apiRoutes.Group("/users", middleware.AuthorizeJWT())
	{
		userSecuredRoutes.GET("/", userHandler.GetAllUsers)
		userSecuredRoutes.GET("/:user_id", userHandler.GetUser)
		adminRoutes := userSecuredRoutes.Group("/", middleware.CheckAdmin())
		{
			adminRoutes.PUT("/:user_id", userHandler.UpdateUser)
			adminRoutes.DELETE("/:user_id", userHandler.DeleteUser)
		}
	}

	productRoutes := apiRoutes.Group("/products", middleware.AuthorizeJWT())
	{
		productRoutes.GET("/", productHandler.GetAllProduct)
		productRoutes.GET("/:product_id", productHandler.GetProduct)
		adminRoutes := productRoutes.Group("/", middleware.CheckAdmin())
		{
			adminRoutes.POST("/", productHandler.CreateProduct)
			adminRoutes.PUT("/:product_id", productHandler.UpdateProduct)
			adminRoutes.DELETE("/:product_id", productHandler.DeleteProduct)
		}
	}

	orderRoutes := apiRoutes.Group("/order", middleware.AuthorizeJWT())
	{
		orderRoutes.POST("/", orderHandler.OrderProducts)
		orderRoutes.GET("/", orderHandler.GetOrders)
		adminRoutes := orderRoutes.Group("/", middleware.CheckAdmin())
		{
			adminRoutes.GET("/:order_id", orderHandler.GetOrderByID)
			adminRoutes.GET("/:order_id/order_items", orderHandler.GetOrderItems)
			adminRoutes.PUT("/", orderHandler.UpdateOrder)
			adminRoutes.PUT("/order_status/:order_id/:status", orderHandler.UpdateOrderStatus)
			adminRoutes.DELETE("/", orderHandler.DeleteOrder)
			adminRoutes.DELETE("/order_items/:order_item_id", orderHandler.DeleteOrderItem)
		}
	}

	return r.Run(address)
}
