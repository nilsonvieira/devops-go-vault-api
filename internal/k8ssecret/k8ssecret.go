package k8ssecret

import (
	"encoding/base64"
	"fmt"
)

func DecodeSecret(data map[string]string) (map[string]string, error) {
	decodedData := make(map[string]string)
	for key, value := range data {
		decodedValue, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("Falha ao decodificar o valor Base64 %s: %v", key, err)
		}
		decodedData[key] = string(decodedValue)
	}
	return decodedData, nil
}
