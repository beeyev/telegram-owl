// Package util contains payload conversion shared by Telegram method packages.
package util

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StructToFormPayload converts a JSON-shaped value into Telegram multipart form
// fields. JSON tags and omitempty rules define field names and presence; boolean
// values use Telegram's conventional "1" and "0" representation.
func StructToFormPayload(v any) (map[string]string, error) {
	if v == nil {
		return nil, errors.New("input struct is nil")
	}

	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}

	var intermediateMap map[string]any
	err = json.Unmarshal(jsonData, &intermediateMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal struct: %w", err)
	}

	result := make(map[string]string, len(intermediateMap))

	for field, val := range intermediateMap {
		switch v := val.(type) {
		case bool:
			if v {
				result[field] = "1"
			} else {
				result[field] = "0"
			}
		default:
			result[field] = fmt.Sprintf("%v", val)
		}
	}

	return result, nil
}
