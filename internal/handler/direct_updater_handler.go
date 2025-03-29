package handler

import (
	"bytes"
	"devops-go-vault-api/internal/vault"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type PasswordUpdateRequest struct {
	BasePath    string `json:"base_path"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	Mode        string `json:"mode,omitempty"`
}

type PasswordUpdateResponse struct {
	Success bool                         `json:"success"`
	Message string                       `json:"message,omitempty"`
	Mode    string                       `json:"mode"`
	Updates []vault.PasswordUpdateResult `json:"updates,omitempty"`
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler corpo da requisição", http.StatusInternalServerError)
		return
	}

	var req PasswordUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Erro ao decodificar a solicitação JSON", http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" {
		http.Error(w, "Senha antiga é obrigatória", http.StatusBadRequest)
		return
	}

	if strings.ToLower(req.Mode) == "edit" && req.NewPassword == "" {
		http.Error(w, "Nova senha é obrigatória no modo edit", http.StatusBadRequest)
		return
	}

	if req.BasePath == "" {
		req.BasePath = "secret"
	}

	if req.Mode == "" {
		req.Mode = "list"
	}

	req.BasePath = strings.TrimSuffix(req.BasePath, "/")

	var mode vault.OperationMode
	switch strings.ToLower(req.Mode) {
	case "list":
		mode = vault.ListMode
	case "edit":
		mode = vault.EditMode
	default:
		http.Error(w, "Modo inválido. Use 'list' ou 'edit'", http.StatusBadRequest)
		return
	}

	oldStdout := os.Stdout
	r1, w1, _ := os.Pipe()
	os.Stdout = w1

	updates, err := vault.SearchAndReplacePasswordDirect(req.BasePath, req.OldPassword, req.NewPassword, mode)

	w1.Close()
	os.Stdout = oldStdout

	var logBuffer bytes.Buffer
	io.Copy(&logBuffer, r1)
	log.Print(logBuffer.String())

	response := PasswordUpdateResponse{
		Mode: req.Mode,
	}

	if err != nil {
		response.Success = false
		response.Message = fmt.Sprintf("Erro ao processar a solicitação: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Updates = updates
		if len(updates) == 0 {
			response.Message = "Nenhuma senha correspondente encontrada"
		} else {
			if mode == vault.ListMode {
				response.Message = fmt.Sprintf("Encontradas %d ocorrências da senha (modo: apenas listagem)", len(updates))
			} else {
				response.Message = fmt.Sprintf("Atualizadas %d ocorrências da senha", len(updates))
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
