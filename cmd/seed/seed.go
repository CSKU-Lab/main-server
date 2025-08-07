package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/sqlx"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	userRepo := sqlx.NewSqlxUserRepository(db)
	userPasswordRepo := sqlx.NewUserPasswordRepository(db)
	userService := services.NewUserService(userRepo, userPasswordRepo)

	user, err := userService.Create(context.Background(), &requests.CreateMultiTypeUser{
		BaseUser: requests.BaseUser{
			Username:    "test_user_1",
			DisplayName: "Test user",
			Roles:       []string{"student"},
		},
		Type: models.UserTypeCredential.String(),
		Password: func() *string {
			password := "test_user_1_password"
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
