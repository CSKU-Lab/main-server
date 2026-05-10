package auth

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndVerifySignedState(t *testing.T) {
	secret := "test-secret-key"

	state, err := generateSignedState(secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, state)

	assert.True(t, verifySignedState(state, secret))
}

func TestVerifySignedState_Tampered(t *testing.T) {
	secret := "test-secret-key"

	state, err := generateSignedState(secret)
	assert.NoError(t, err)

	tampered := state + "x"
	assert.False(t, verifySignedState(tampered, secret))
}

func TestVerifySignedState_WrongSecret(t *testing.T) {
	secret := "test-secret-key"

	state, err := generateSignedState(secret)
	assert.NoError(t, err)

	assert.False(t, verifySignedState(state, "wrong-secret"))
}

func TestVerifySignedState_Expired(t *testing.T) {
	secret := "test-secret-key"

	// Manually craft an expired state
	payload := "test-uuid:" + fmt.Sprintf("%d", time.Now().Add(-20*time.Minute).Unix())
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := hmacSig(encodedPayload, secret)
	encodedSig := base64.RawURLEncoding.EncodeToString(sig)
	expiredState := encodedPayload + "." + encodedSig

	assert.False(t, verifySignedState(expiredState, secret))
}
