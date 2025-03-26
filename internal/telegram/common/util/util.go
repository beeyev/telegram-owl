package util

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StructToFormPayload parses parameter struct to key value structure, v should be a pointer to struct.
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

	// Convert the map values to strings
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
