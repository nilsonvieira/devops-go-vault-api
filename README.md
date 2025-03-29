# DevOps Vault API

Uma API em Go para gerenciamento avançado de segredos no HashiCorp Vault. Esta ferramenta permite diversas operações relacionadas a segredos, incluindo armazenamento, conversão de formatos, decodificação, gerenciamento de credenciais de banco de dados e busca recursiva com substituição de senhas.

## Funcionalidades

- **Armazenamento de Segredos**: Armazene dados estruturados no Vault
- **Conversão de Formatos**: Converta YAMLs para diferentes formatos compatíveis com o Vault
- **Decodificação de Segredos**: Decodifique segredos base64 do Kubernetes
- **Gerenciamento de Credenciais de DB**: Geração de estruturas específicas para credenciais de banco de dados
- **Deleção de Segredos**: Remova segredos de forma segura do Vault
- **Conversão JSON→Vault**: Transforme estruturas JSON para o formato do Vault
- **Busca e Substituição Recursiva**: Encontre e substitua senhas específicas em toda a estrutura de segredos do Vault

## Requisitos

- Go 1.22+
- HashiCorp Vault 1.16.1+
- Token de acesso ao Vault com permissões adequadas

## Instalação e Configuração

### Usando Go

1. Clone o repositório:
   ```bash
   git clone https://github.com/seu-usuario/devops-go-vault-api.git
   cd devops-go-vault-api
   ```

2. Instale as dependências:
   ```bash
   go mod download
   ```

3. Configure as variáveis de ambiente:
   ```bash
   cp env_template .env
   # Edite o arquivo .env com suas configurações
   ```

4. Execute a aplicação:
   ```bash
   go run cmd/server/main.go
   ```

### Usando Docker

1. Construa a imagem:
   ```bash
   docker build -t devops-go-vault-api .
   ```

2. Execute o container:
   ```bash
   docker run -p 8080:8080 --env-file .env devops-go-vault-api
   ```

## Endpoints da API

### 1. Armazenar Dados no Vault

**Endpoint:** `POST /sendVault`

Armazena um ou mais segredos no Vault.

**Corpo da requisição:**
```json
[
   {
      "path": "secret/data/meu-servico/config",
      "data": {
         "api_key": "chave123",
         "ambiente": "producao",
         "timeout": "30s"
      }
   }
]
```

### 2. Converter YAML para Diferentes Formatos

**Endpoint:** `POST /convert`

Converte um YAML para diversos formatos úteis em diferentes contextos.

**Corpo da requisição (formato YAML):**
```yaml
database:
   username: admin
   password: senha123
   host: db.exemplo.com
   port: 5432
```

### 3. Decodificar Segredos do Kubernetes

**Endpoint:** `POST /decSecret`

Decodifica segredos do Kubernetes que estão em formato Base64.

**Corpo da requisição (formato YAML):**
```yaml
data:
   username: YWRtaW4=
   password: c2VuaGExMjM=
```

### 4. Gerar Estruturas para Banco de Dados

**Endpoint:** `POST /generate`

Gera estruturas de configuração para diferentes tipos de banco de dados.

**Corpo da requisição:**
```json
{
   "dbInfo": {
      "username": "admin",
      "password": "senha123",
      "host": "db.exemplo.com",
      "port": "5432"
   },
   "host": "db.exemplo.com",
   "sgbd": "postgres",
   "application": "meu-app"
}
```

### 5. Deletar Segredos

**Endpoint:** `DELETE /deleteSecret`

Remove segredos do Vault.

**Corpo da requisição:**
```json
{
   "path": "secret/data/meu-servico/config"
}
```

### 6. Converter JSON para Formato Vault

**Endpoint:** `POST /jsonToVaultJson`

Converte configurações de bancos de dados para o formato do Vault.

**Corpo da requisição:**
```json
[
   {
      "POSTGRES_HOST": "db.exemplo.com",
      "POSTGRES_PORT": "5432",
      "POSTGRES_USERNAME": "admin",
      "POSTGRES_PASSWORD": "senha123"
   }
]
```

**Parâmetros de query:**
- `application`: Nome da aplicação (ex: `meu-app`)

### 7. Buscar e Atualizar Senhas Recursivamente

**Endpoint:** `POST /updatePassword`

Busca recursivamente no Vault por senhas específicas e, dependendo do modo, lista ou substitui as ocorrências. A API navega automaticamente por toda a estrutura de diretórios do Vault, identifica o formato correto dos segredos e processa as senhas que correspondam exatamente ao valor especificado.

**Corpo da requisição:**
```json
{
  "base_path": "secret/caminho/base",
  "old_password": "senha-antiga-exata",
  "new_password": "nova-senha-segura",
  "mode": "list"
}
```

**Parâmetros:**
- `base_path`: O caminho base para iniciar a busca (ex: "secret/minha-app")
- `old_password`: A senha exata que você deseja encontrar
- `new_password`: A nova senha a ser aplicada (obrigatório apenas no modo "edit")
- `mode`: O modo de operação (padrão: "list")
   - `list`: Apenas lista as ocorrências sem fazer alterações
   - `edit`: Encontra e substitui as ocorrências pela nova senha

**Exemplo de resposta em modo "list":**
```json
{
  "success": true,
  "message": "Encontradas 3 ocorrências da senha (modo: apenas listagem)",
  "mode": "list",
  "updates": [
    {
      "path": "secret/data/minha-app/segredo1",
      "key": "password"
    },
    {
      "path": "secret/data/minha-app/subpasta/segredo2",
      "key": "api_key"
    },
    {
      "path": "secret/data/minha-app/config/db",
      "key": "senha"
    }
  ]
}
```

**Exemplo de resposta em modo "edit":**
```json
{
  "success": true,
  "message": "Atualizadas 3 ocorrências da senha",
  "mode": "edit",
  "updates": [
    {
      "path": "secret/data/minha-app/segredo1",
      "key": "password"
    },
    {
      "path": "secret/data/minha-app/subpasta/segredo2",
      "key": "api_key"
    },
    {
      "path": "secret/data/minha-app/config/db",
      "key": "senha"
    }
  ]
}
```

## Exemplo de Uso com cURL

### Listar ocorrências de uma senha sem alterar:
```bash
curl -X POST http://localhost:8080/updatePassword \
  -H "Content-Type: application/json" \
  -d '{
    "base_path": "secret/minha-app",
    "old_password": "senha-antiga",
    "mode": "list"
  }'
```

### Buscar e substituir senhas:
```bash
curl -X POST http://localhost:8080/updatePassword \
  -H "Content-Type: application/json" \
  -d '{
    "base_path": "secret/minha-app",
    "old_password": "senha-antiga",
    "new_password": "nova-senha-segura",
    "mode": "edit"
  }'
```

## Considerações de Segurança

- **HTTPS**: Configure TLS/HTTPS para proteger as comunicações entre o cliente e a API
- **Gestão de Tokens**: Use variáveis de ambiente para armazenar credenciais do Vault
- **Políticas de Acesso**: Configure políticas adequadas no Vault para limitar o acesso
- **Logs**: Implemente logs detalhados para auditoria de operações sensíveis
- **Modo List**: Use o modo "list" para verificação antes de fazer alterações em produção

## Estrutura do Projeto

```
.
├── cmd
│   └── server
│       └── main.go                # Ponto de entrada da aplicação
├── config
│   └── config.go                 # Carregamento de configurações
├── internal
│   ├── converter
│   │   └── converter.go          # Conversão de formatos YAML
│   ├── handler
│   │   ├── handler.go            # Handlers da API
│   │   └── direct_updater_handler.go # Handler de atualização de senhas
│   ├── k8ssecret
│   │   └── k8ssecret.go          # Decodificação de segredos K8s
│   └── vault
│       ├── vault.go              # Operações básicas do Vault
│       └── direct_updater.go     # Busca e substituição de senhas
├── .gitignore
├── Dockerfile
├── env_template                  # Template para variáveis de ambiente
├── go.mod
├── go.sum
└── README.md
```

## Licença

Este projeto está licenciado sob a [Licença MIT](LICENSE).