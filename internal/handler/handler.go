package handler

import (
	"devops-go-vault-api"
	"devops-go-vault-api/internal/converter"
	"devops-go-vault-api/internal/k8ssecret"
	"devops-go-vault-api/internal/vault"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"strings"
)

type Request struct {
	Path string            `json:"path"`
	Data map[string]string `json:"data"`
}

type SecretRequest struct {
	Data map[string]string `json:"data"`
}

type DBInfo struct {
	Data map[string]string `json:"data"`
}

type GenerateRequest struct {
	DBInfo      map[string]string `json:"dbInfo"`
	Host        string            `json:"host"`
	SGBD        string            `json:"sgbd"`
	Application string            `json:"application"`
}

type DeleteSecretRequest struct {
	Path string `json:"path"`
}

type DBConfig struct {
	DB       string `json:"DB"`
	Host     string `json:"HOST"`
	Password string `json:"PASSWORD"`
	Port     string `json:"PORT"`
	Username string `json:"USERNAME"`
}

type SecretPayload struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

func StoreHandler(w http.ResponseWriter, r *http.Request) {
	var requests []Request
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, req := range requests {
		if req.Path == "" || len(req.Data) == 0 {
			http.Error(w, "Path and Data são necessários", http.StatusBadRequest)
			return
		}

		err = vault.StoreInVault(req.Path, req.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Estrutura criada e Dados Inseridos com Sucesso!")
}

func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flatMap, upperMap, templateMap, err := converter.FlattenYAML(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"output1": flatMap,
		"output2": upperMap,
		"output3": templateMap,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func DecryptSecretHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req SecretRequest
	err = yaml.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Data) == 0 {
		http.Error(w, "Dados Requeridos!", http.StatusBadRequest)
		return
	}

	decodedData, err := k8ssecret.DecodeSecret(req.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(decodedData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func JsonHandler(w http.ResponseWriter, r *http.Request) {
	config, err := devops_go_vault_api.ParseIniFile("config.ini")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json1 := devops_go_vault_api.TransformToJson1(config)
	json2 := devops_go_vault_api.TransformToJson2(config)

	result := map[string]interface{}{
		"output1": json1,
		"output2": json2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Host == "" || req.SGBD == "" || req.Application == "" || len(req.DBInfo) == 0 {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	lowerSGBD := strings.ToLower(req.SGBD) // SGBD em minúsculas para a URL
	upperSGBD := strings.ToUpper(req.SGBD) // SGBD em maiúsculas para o prefixo das chaves
	templatePath := fmt.Sprintf("secret/data/general/dba/%s/%s/%s", lowerSGBD, req.Host, req.Application)

	templateOutput := make(map[string]string)
	for key := range req.DBInfo {
		upperKey := strings.ToUpper(key)
		templateKey := fmt.Sprintf("{{%s::%s_%s}}", templatePath, upperSGBD, upperKey)
		templateOutput[key] = templateKey
	}

	renamedOutput := make(map[string]string)
	for key := range req.DBInfo {
		upperKey := strings.ToUpper(key)
		renamedKey := fmt.Sprintf("%s_%s", upperSGBD, upperKey)
		renamedOutput[renamedKey] = req.DBInfo[key]
	}

	response := map[string]interface{}{
		"template_output": templateOutput,
		"renamed_output":  renamedOutput,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func DeleteSecretHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecretRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil || req.Path == "" {
		http.Error(w, "Invalid or missing path", http.StatusBadRequest)
		return
	}

	secretList, err := vault.ListSecrets(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(secretList) > 0 {
		http.Error(w, "Cannot delete a folder with multiple secrets", http.StatusForbidden)
		return
	}

	err = vault.DeleteSecret(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": fmt.Sprintf("Secret at path '%s' deleted successfully", req.Path),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func GenerateSecretHandler(w http.ResponseWriter, r *http.Request) {
	var dbConfigs []map[string]string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &dbConfigs)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	application := r.URL.Query().Get("application")
	if application == "" {
		http.Error(w, "Missing application parameter", http.StatusBadRequest)
		return
	}

	var result []SecretPayload
	pathDataMap := make(map[string]map[string]interface{})
	legacyData := make(map[string]interface{})

	for _, config := range dbConfigs {
		var dbType, port, host string
		for key, value := range config {
			if strings.HasSuffix(key, "_PORT") {
				port = value
			}
			if strings.HasSuffix(key, "_HOST") {
				host = value
			}
		}

		switch port {
		case "1433":
			dbType = "sqlserver"
		case "49600":
			dbType = "sqlserver"
		case "5432":
			dbType = "postgres"
		case "1521":
			dbType = "oracle"
		default:
			http.Error(w, "Unsupported port", http.StatusBadRequest)
			return
		}

		upperDBType := strings.ToUpper(dbType)

		path := fmt.Sprintf("secret/data/general/dba/%s/%s/%s", dbType, host, application)

		if _, exists := pathDataMap[path]; !exists {
			pathDataMap[path] = make(map[string]interface{})
		}

		for key, value := range config {
			upperKey := fmt.Sprintf("%s_%s", upperDBType, strings.ToUpper(key))
			pathDataMap[path][upperKey] = value

			legacyKey := strings.TrimPrefix(upperKey, upperDBType+"_")
			legacyData[legacyKey] = fmt.Sprintf("{{%s::%s}}", path, upperKey)
		}
	}

	for path, data := range pathDataMap {
		result = append(result, SecretPayload{
			Path: path,
			Data: data,
		})
	}

	result = append(result, SecretPayload{
		Path: "secret/data/legacy/" + application,
		Data: legacyData,
	})

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
