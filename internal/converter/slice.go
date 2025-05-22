package converter

import "fmt"

func ToStringSlice(slice []any) ([]string, error) {
	result := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		result[i] = str
	}
	return result, nil
}
