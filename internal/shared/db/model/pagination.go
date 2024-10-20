package dbmodel

import (
	"encoding/base64"
	"encoding/json"
)

func DecodePaginationToken[T any](token string) (*T, error) {
	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var d T
	err = json.Unmarshal(tokenBytes, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func GetPaginationToken[PaginationType any](d PaginationType) (string, error) {
	tokenBytes, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
