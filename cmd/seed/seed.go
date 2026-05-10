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
	"github.com/CSKU-Lab/main-server/domain/repositories"
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

type SeedResult struct {
	Password string
	Skipped  bool
}

func runSeed(ctx context.Context, userService services.UserService, userGroupService services.UserGroupService, userGroupRepo repositories.UserGroup) (*SeedResult, error) {
	password, err := generatePassword(20)
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}

	groupID, err := userGroupService.Create(ctx, "Administrators")
	if err != nil {
		csErr, ok := err.(*cserrors.Error)
		if !ok || csErr.HttpStatus != http.StatusConflict {
			return nil, fmt.Errorf("create user group: %w", err)
		}
		existing, err := userGroupRepo.GetByName(ctx, "Administrators")
		if err != nil {
			return nil, fmt.Errorf("fetch existing user group: %w", err)
		}
		groupID = existing.ID
	}

	plainPassword := password
	err = userService.Create(ctx, &requests.CreateMultiTypeUser{
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
			return &SeedResult{Skipped: true}, nil
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &SeedResult{Password: plainPassword}, nil
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

	result, err := runSeed(context.Background(), userService, userGroupService, userGroupRepo)
	if err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	if result.Skipped {
		fmt.Println("ℹ️ User already exists, skipping")
		return
	}

	fmt.Println("✅ Seed completed")
	fmt.Println("==========================================")
	fmt.Println("  Username:", "admin")
	fmt.Println("  Password:", result.Password)
	fmt.Println("==========================================")
}
