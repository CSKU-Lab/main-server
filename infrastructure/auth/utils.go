package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const stateExpiry = 10 * time.Minute

func generateSignedState(secret string) (string, error) {
	gen, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	payload := fmt.Sprintf("%s:%d", gen.String(), time.Now().Unix())
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))

	sig := hmacSig(encodedPayload, secret)
	encodedSig := base64.RawURLEncoding.EncodeToString(sig)

	return encodedPayload + "." + encodedSig, nil
}

func verifySignedState(state, secret string) bool {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return false
	}

	encodedPayload := parts[0]
	encodedSig := parts[1]

	expectedSig := hmacSig(encodedPayload, secret)
	decodedSig, err := base64.RawURLEncoding.DecodeString(encodedSig)
	if err != nil {
		return false
	}

	if !hmac.Equal(expectedSig, decodedSig) {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return false
	}

	payload := string(payloadBytes)
	parts = strings.Split(payload, ":")
	if len(parts) != 2 {
		return false
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	if time.Since(time.Unix(ts, 0)) > stateExpiry {
		return false
	}

	return true
}

func hmacSig(data, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}
