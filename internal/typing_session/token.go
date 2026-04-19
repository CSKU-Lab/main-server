package typing_session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
)

const maxSessionAge = 30 * time.Minute

type TokenClaims struct {
	StudentID  string    `json:"student_id"`
	MaterialID string    `json:"material_id"`
	LabID      string    `json:"lab_id"`
	StartedAt  time.Time `json:"started_at"`
}

func GenerateToken(secret string, claims *TokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, encodedPayload)
	return encodedPayload + "." + sig, nil
}

func VerifyToken(secret, token string) (*TokenClaims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid session token"})
	}

	encodedPayload, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sig), []byte(sign(secret, encodedPayload))) {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid session token"})
	}

	raw, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid session token"})
	}

	var claims TokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid session token"})
	}

	if time.Since(claims.StartedAt) > maxSessionAge {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Session token expired"})
	}

	return &claims, nil
}

func sign(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
