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
	uowRepo := sqlx.NewUoWRepository(context.Background(), db)
	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)
	userGroupService := services.NewUserGroupService(userGroupRepo)

	// Create "Postman Users" group if it doesn't exist
	id, err := userGroupService.Create(context.Background(), "Postman Users")
	if err != nil {
		fmt.Println("❌ Error creating user group:", err)
		os.Exit(1)
	}

	// Create admin user
	err = userService.Create(context.Background(), &requests.CreateMultiTypeUser{
		Username:    "postman_admin",
		DisplayName: "Postman Admin",
		Roles:       []string{"admin"},

		GroupID: &id,
		Type:    models.UserTypeCredential,
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
}
