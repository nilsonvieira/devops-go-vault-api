package vault

import (
	"devops-go-vault-api/config"
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
