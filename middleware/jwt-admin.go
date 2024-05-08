package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go_final/models"
	"net/http"
)

func CheckAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, adminExists := ctx.Get("role")
		if adminExists && role.(string) == string(models.ADMIN_ROLE) {
			ctx.Next()
			return
		}

		id := ctx.Param("user_id")
		if ctxID, ok := ctx.Get("userID"); ok {
			ctxIDStr := fmt.Sprintf("%v", ctxID)
			if ctxIDStr == id {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
	}
}
