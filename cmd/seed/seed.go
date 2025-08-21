package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	userRepo := sqlx.NewUserRepository(db)
	userPasswordRepo := sqlx.NewUserPasswordRepository(db)
	userGroupRepo := sqlx.NewUserGroupRepository(db)
	uowRepo := sqlx.NewUserUoWRepository(context.Background(), db)
	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)

	user, err := userService.Create(context.Background(), &requests.CreateMultiTypeUser{
		BaseUser: requests.BaseUser{
			Username:    "postman_admin",
			DisplayName: "Postman Admin",
			Roles:       []string{"admin"},
		},
		GroupID: func() *string {
			groupID := "0198b246-1023-7486-a358-3a24ea6c3435"
			return &groupID
		}(),
		Type: models.UserTypeCredential.String(),
		Password: func() *string {
			password := "postman_admin"
			return &password
		}(),
	})
	if err != nil {
		fmt.Println("❌ Error creating user:", err)
		os.Exit(1)
	}

	fmt.Println("✅ User created successfully")
	fmt.Println(user)
}
