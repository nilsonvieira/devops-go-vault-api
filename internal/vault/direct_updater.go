package vault

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

type PasswordUpdateResult struct {
	Path  string `json:"path"`
	Key   string `json:"key"`
	Error string `json:"error,omitempty"`
}

type OperationMode string

const (
	ListMode OperationMode = "list"
	EditMode OperationMode = "edit"
)

func SearchAndReplacePasswordDirect(basePath, oldPassword, newPassword string, mode OperationMode) ([]PasswordUpdateResult, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	if mode != ListMode && mode != EditMode {
		return nil, fmt.Errorf("modo de operação inválido: %s (use 'list' ou 'edit')", mode)
	}

	if mode == ListMode {
		fmt.Printf("Executando em modo LISTA: apenas buscando ocorrências sem fazer alterações\n")
	} else {
		fmt.Printf("Executando em modo EDIÇÃO: buscando e substituindo ocorrências\n")
	}

	var allUpdates []PasswordUpdateResult
	var processedPaths = make(map[string]bool)

	basePath = strings.TrimSuffix(basePath, "/")
	basePath = normalizePathSlashes(basePath)

	baseVariations := generatePathVariations(basePath)

	for _, variant := range baseVariations {
		variant = normalizePathSlashes(variant)

		list, err := client.Logical().List(variant)
		if err != nil {
			continue
		}

		if list == nil || list.Data == nil {
			continue
		}

		keys, ok := list.Data["keys"].([]interface{})
		if !ok {
			continue
		}

		fmt.Printf("Encontrados %d itens em %s\n", len(keys), variant)

		for _, keyRaw := range keys {
			key, ok := keyRaw.(string)
			if !ok {
				continue
			}

			if strings.HasSuffix(key, "/") {
				childPath := normalizePathSlashes(fmt.Sprintf("%s/%s", variant, key))
				childPath = strings.TrimSuffix(childPath, "/")

				fmt.Printf("Explorando subdiretório: %s\n", childPath)

				childUpdates, _ := SearchAndReplacePasswordDirect(childPath, oldPassword, newPassword, mode)
				for _, update := range childUpdates {
					uniqueKey := update.Path + ":" + update.Key
					if !processedPaths[uniqueKey] {
						allUpdates = append(allUpdates, update)
						processedPaths[uniqueKey] = true
					}
				}
			} else {
				secretPath := normalizePathSlashes(fmt.Sprintf("%s/%s", variant, key))
				secretVariations := generateSecretPathVariations(secretPath)

				fmt.Printf("Verificando segredo: %s\n", key)

				for _, secretVariant := range secretVariations {
					secretVariant = normalizePathSlashes(secretVariant)
					updates := processSecret(client, secretVariant, oldPassword, newPassword, mode)
					if len(updates) > 0 {
						for _, update := range updates {
							uniqueKey := update.Path + ":" + update.Key
							if !processedPaths[uniqueKey] {
								allUpdates = append(allUpdates, update)
								processedPaths[uniqueKey] = true
							}
						}
						break
					}
				}
			}
		}
	}

	secretVariations := generateSecretPathVariations(basePath)
	for _, secretVariant := range secretVariations {
		secretVariant = normalizePathSlashes(secretVariant)
		updates := processSecret(client, secretVariant, oldPassword, newPassword, mode)
		if len(updates) > 0 {
			for _, update := range updates {
				uniqueKey := update.Path + ":" + update.Key
				if !processedPaths[uniqueKey] {
					allUpdates = append(allUpdates, update)
					processedPaths[uniqueKey] = true
				}
			}
			break
		}
	}

	return allUpdates, nil
}

func normalizePathSlashes(path string) string {
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func generatePathVariations(path string) []string {
	variations := []string{
		path,
		path + "/",
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		metadataPath := fmt.Sprintf("%s/metadata/%s", parts[0], parts[1])
		variations = append(variations, metadataPath)
		variations = append(variations, metadataPath+"/")
	}

	return variations
}

func generateSecretPathVariations(path string) []string {
	variations := []string{
		path,
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		if strings.Contains(path, "/metadata/") {
			dataPath := strings.Replace(path, "/metadata/", "/data/", 1)
			variations = append(variations, dataPath)
		} else if !strings.Contains(path, "/data/") {
			dataPath := fmt.Sprintf("%s/data/%s", parts[0], parts[1])
			variations = append(variations, dataPath)
		}
	}

	return variations
}

func processSecret(client *api.Client, path string, oldPassword, newPassword string, mode OperationMode) []PasswordUpdateResult {
	var updates []PasswordUpdateResult

	path = normalizePathSlashes(path)

	secret, err := client.Logical().Read(path)
	if err != nil {
		return updates
	}

	if secret == nil || secret.Data == nil {
		return updates
	}

	var dataMap map[string]interface{}
	var format string

	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		format = "KV v2"
		dataMap = data
	} else {
		format = "KV v1 ou direto"
		dataMap = secret.Data
	}

	var updated bool
	updatedData := make(map[string]interface{})

	for key, value := range dataMap {
		updatedData[key] = value

		if strValue, ok := value.(string); ok {
			if strValue == oldPassword {
				if mode == EditMode {
					fmt.Printf("SENHA ENCONTRADA em %s, campo: %s (será alterada)\n", path, key)

					updatedData[key] = newPassword
					updated = true
				} else {
					fmt.Printf("SENHA ENCONTRADA em %s, campo: %s (modo lista: sem alteração)\n", path, key)
				}

				updates = append(updates, PasswordUpdateResult{
					Path: path,
					Key:  key,
				})
			}
		}
	}

	if updated && mode == EditMode {
		var payload map[string]interface{}

		if format == "KV v2" {
			payload = map[string]interface{}{
				"data": updatedData,
			}
		} else {
			payload = updatedData
		}

		_, err := client.Logical().Write(path, payload)
		if err != nil {
			fmt.Printf("ERRO ao atualizar segredo %s: %v\n", path, err)

			lastIdx := len(updates) - 1
			updates[lastIdx].Error = err.Error()
		} else {
			fmt.Printf("Atualização bem-sucedida em %s\n", path)
		}
	}

	return updates
}
