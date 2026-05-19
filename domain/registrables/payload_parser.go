package registrables

import (
	"encoding/json"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
)

func parsePayload[T any](bytesPayload []byte) (*T, error) {
	var unmarshal struct {
		Payload *T `json:"payload"`
	}

	err := json.Unmarshal(bytesPayload, &unmarshal)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			Message:    "invalid payload format",
			HttpStatus: http.StatusBadRequest,
		})
	}

	return unmarshal.Payload, nil
}

func buildPayloadRequest(payload any) ([]byte, error) {
	rawReq, err := json.Marshal(struct {
		Payload any `json:"payload"`
	}{
		Payload: payload,
	})
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			Message:    "failed to build material payload",
			HttpStatus: http.StatusInternalServerError,
		})
	}

	return rawReq, nil
}
