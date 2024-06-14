package handler

import (
	"devops-go-vault-api/internal/converter"
	"devops-go-vault-api/internal/vault"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Path string            `json:"path"`
	Data map[string]string `json:"data"`
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
