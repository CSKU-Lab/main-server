package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/sqlx"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	userRepo := sqlx.NewSqlxUserRepository(db)
	userService := services.NewUserService(userRepo)

	user, err := userService.Create(context.Background(), &requests.User{
		Username:    "SornchaiTheDev",
		Email:       "sornchaithedev@gmail.com",
		DisplayName: "Sornchai Somsakul",
		Roles:       []string{"admin"},
	})
	if err != nil {
		fmt.Println("❌ Error creating user:", err)
		os.Exit(1)
	}

	fmt.Println("✅ User created successfully")
	fmt.Println(user)
}
