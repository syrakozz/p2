package deepgram

import (
	"encoding/json"

	"disruptive/lib/common"
)

type deepgramError struct {
	Code      string `json:"err_code"`
	Message   string `json:"err_msg"`
	RequestID string `json:"request_id"`
}

func convertDeepgramError(deepgramErr error) error {
	dErr := deepgramError{}

	err := json.Unmarshal([]byte(deepgramErr.Error()), &dErr)
	if err != nil {
		return deepgramErr
	}

	switch dErr.Code {
	case "INVALID_AUTH", "INSUFFICIENT_PERMISSIONS":
		if dErr.Message == "Project does not have access to the requested model." {
			return common.ErrForbidden{Err: deepgramErr, Msg: dErr.Message}
		}
		return common.ErrUnauthorized
	case "PROJECT_NOT_FOUND":
		return common.ErrNotFound{Err: deepgramErr, Msg: dErr.Message}
	case "ASR_PAYMENT_REQUIRED":
		return common.ErrPaymentRequired{Err: deepgramErr, Msg: dErr.Message}
	case "INTERNAL_SERVER_ERROR":
		return common.ErrConnection{Err: deepgramErr, Msg: dErr.Message}
	default:
		return deepgramErr
	}
}
