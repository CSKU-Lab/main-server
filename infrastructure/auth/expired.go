package auth

import "time"

var (
	ACCESS_TOKEN_EXPIRED_TIME  = time.Now().Add(time.Hour * 1)
	REFRESH_TOKEN_EXPIRED_TIME = time.Now().Add(time.Hour * 24 * 5)
)
