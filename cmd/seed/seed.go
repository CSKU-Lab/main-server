package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	config := configs.NewConfig()
	db := configs.NewDB(config)

	userRepo := sqlx.NewUserRepository(db)
	userPasswordRepo := sqlx.NewUserPasswordRepository(db)
	userGroupRepo := sqlx.NewUserGroupRepository(db)
	uowRepo := sqlx.NewUoWRepository(context.Background(), db)
	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)
	userGroupService := services.NewUserGroupService(userGroupRepo)

	username := getEnv("SEED_USERNAME", "seed_admin")
	displayName := getEnv("SEED_DISPLAY_NAME", "Seed Admin")
	rolesRaw := getEnv("SEED_ROLES", "admin")
	roles := strings.Split(rolesRaw, ",")
	userType := models.UserType(getEnv("SEED_TYPE", "credential"))

	req := &requests.CreateMultiTypeUser{
		Username:    username,
		DisplayName: displayName,
		Roles:       roles,
		Type:        userType,
	}

	switch userType {
	case models.UserTypeCredential:
		groupName := getEnv("SEED_GROUP", "Seed Users")
		password := os.Getenv("SEED_PASSWORD")
		if password == "" {
			fmt.Println("❌ SEED_PASSWORD is required for credential user")
			os.Exit(1)
		}

		id, err := userGroupService.Create(context.Background(), groupName)
		if err != nil {
			fmt.Println("❌ Error creating user group:", err)
			os.Exit(1)
		}
		req.GroupID = &id
		req.Password = &password

	case models.UserTypeOauth:
		email := os.Getenv("SEED_EMAIL")
		if email == "" {
			fmt.Println("❌ SEED_EMAIL is required for oauth user")
			os.Exit(1)
		}
		req.Email = &email

	default:
		fmt.Printf("❌ Unknown SEED_TYPE: %q (must be 'credential' or 'oauth')\n", userType)
		os.Exit(1)
	}

	err := userService.Create(context.Background(), req)
	if err != nil {
		fmt.Println("❌ Error creating user:", err)
		os.Exit(1)
	}

	fmt.Println("✅ User created successfully")
}
