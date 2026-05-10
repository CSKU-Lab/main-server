//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDatabaseURL() string {
	host := getEnvOrDefault("PGHOST", "localhost")
	port := getEnvOrDefault("PGPORT", "5432")
	user := getEnvOrDefault("PGUSER", "cs_pg_user")
	password := getEnvOrDefault("PGPASSWORD", "cs_pg_password")
	dbname := getEnvOrDefault("PGDATABASE", "main-server")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("postgres", getTestDatabaseURL())
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func buildServices(db *sqlx.DB) (services.UserService, services.UserGroupService, repositories.UserGroup) {
	userRepo := sqlxAdapter.NewUserRepository(db)
	userPasswordRepo := sqlxAdapter.NewUserPasswordRepository(db)
	userGroupRepo := sqlxAdapter.NewUserGroupRepository(db)
	uowRepo := sqlxAdapter.NewUoWRepository(context.Background(), db)
	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)
	userGroupService := services.NewUserGroupService(userGroupRepo)
	return userService, userGroupService, userGroupRepo
}

func cleanupSeed(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		DELETE FROM user_passwords WHERE user_id IN (SELECT id FROM users WHERE username = 'admin');
		DELETE FROM users WHERE username = 'admin';
		DELETE FROM user_groups WHERE name = 'Administrators';
	`)
	require.NoError(t, err)
}

func TestRunSeed_FirstRun(t *testing.T) {
	db := setupDB(t)
	cleanupSeed(t, db)
	t.Cleanup(func() { cleanupSeed(t, db) })

	userService, userGroupService, userGroupRepo := buildServices(db)

	result, err := runSeed(context.Background(), userService, userGroupService, userGroupRepo)

	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.NotEmpty(t, result.Password)
	assert.Len(t, result.Password, 20)
}

func TestRunSeed_IdempotentOnSecondRun(t *testing.T) {
	db := setupDB(t)
	cleanupSeed(t, db)
	t.Cleanup(func() { cleanupSeed(t, db) })

	userService, userGroupService, userGroupRepo := buildServices(db)
	ctx := context.Background()

	_, err := runSeed(ctx, userService, userGroupService, userGroupRepo)
	require.NoError(t, err, "first run should succeed")

	result, err := runSeed(ctx, userService, userGroupService, userGroupRepo)
	require.NoError(t, err, "second run should not error")
	assert.True(t, result.Skipped, "second run should skip")
}

func TestRunSeed_AdminUserCreated(t *testing.T) {
	db := setupDB(t)
	cleanupSeed(t, db)
	t.Cleanup(func() { cleanupSeed(t, db) })

	userService, userGroupService, userGroupRepo := buildServices(db)

	_, err := runSeed(context.Background(), userService, userGroupService, userGroupRepo)
	require.NoError(t, err)

	user, err := userService.GetByUsername(context.Background(), "admin")
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "Administrator", user.DisplayName)
	assert.Contains(t, user.Roles, models.Role("admin"))
}
