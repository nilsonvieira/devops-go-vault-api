package vault

import (
	"devops-go-vault-api/config"
	"fmt"
	"github.com/hashicorp/vault/api"
)

func StoreInVault(path string, data map[string]string) error {

	conf := api.DefaultConfig()
	conf.Address = config.VaultAddress

	client, err := api.NewClient(conf)
	if err != nil {
		return err
	}

	client.SetToken(config.VaultToken)

	secretData := map[string]interface{}{
		"data": data,
	}

	_, err = client.Logical().Write(path, secretData)
	return err
}

func getClient() (*api.Client, error) {
	conf := api.DefaultConfig()
	conf.Address = config.VaultAddress // Usa o endereço do Vault do .env

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	client.SetToken(config.VaultToken)

	return client, nil
}

func ListSecrets(path string) ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	secret, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at path '%s': %v", path, err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	if keys, ok := secret.Data["keys"].([]interface{}); ok {
		secretList := []string{}
		for _, key := range keys {
			secretList = append(secretList, key.(string))
		}
		return secretList, nil
	}

	return []string{}, nil
}

func DeleteSecret(path string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metadataPath := fmt.Sprintf("secret/metadata/%s", path)

	_, err = client.Logical().Delete(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to delete secret at path '%s': %v", path, err)
	}

	return nil
}
