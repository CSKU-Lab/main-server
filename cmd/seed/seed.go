package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:length], nil
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

	password, err := generatePassword(20)
	if err != nil {
		fmt.Println("❌ Error generating password:", err)
		os.Exit(1)
	}

	groupID, err := userGroupService.Create(context.Background(), "Administrators")
	if err != nil {
		csErr, ok := err.(*cserrors.Error)
		if !ok || csErr.HttpStatus != http.StatusConflict {
			fmt.Println("❌ Error creating user group:", err)
			os.Exit(1)
		}
		existing, err := userGroupRepo.GetByName(context.Background(), "Administrators")
		if err != nil {
			fmt.Println("❌ Error fetching existing user group:", err)
			os.Exit(1)
		}
		groupID = existing.ID
		fmt.Println("ℹ️ User group already exists, using existing group")
	}

	fmt.Println("✅ Seed completed")
	fmt.Println("==========================================")
	fmt.Println("  Username:", "admin")
	fmt.Println("  Password:", password)
	fmt.Println("==========================================")

	err = userService.Create(context.Background(), &requests.CreateMultiTypeUser{
		Username:    "admin",
		DisplayName: "Administrator",
		Roles:       []string{"admin"},
		GroupID:     &groupID,
		Type:        models.UserTypeCredential,
		Password:    &password,
	})
	if err != nil {
		csErr, ok := err.(*cserrors.Error)
		if ok && csErr.HttpStatus == http.StatusConflict {
			fmt.Println("ℹ️ User already exists, skipping")
			return
		}
		fmt.Println("❌ Error creating user:", err)
		os.Exit(1)
	}
}
