package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CSKU-Lab/main-server/configs"
)

func main() {
	cfg := configs.NewConfig()
	db := configs.NewDB(cfg)

	result, err := db.ExecContext(context.Background(), `
		INSERT INTO user_auth_providers (user_id, provider)
		SELECT id,
		       CASE type
		           WHEN 'credential' THEN 'credential'::auth_provider
		           WHEN 'oauth'       THEN 'google'::auth_provider
		       END
		FROM users
		WHERE is_deleted = false
		ON CONFLICT (user_id, provider) DO NOTHING
	`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Backfill failed:", err)
		os.Exit(1)
	}

	n, _ := result.RowsAffected()
	fmt.Printf("✅ Backfill complete: %d rows inserted into user_auth_providers\n", n)
}
