# Finance App

SaaS de Planejamento Financeiro construido com Go, Echo e PostgreSQL, organizado em Clean Architecture.

## Requisitos

- [Docker](https://www.docker.com/) e Docker Compose

Nao e necessario ter Go instalado localmente. Todo o ambiente roda dentro dos containers.

## Instalacao

### 1. Clone o repositorio

```bash
git clone https://github.com/eduardogarbin/finance-app.git
cd finance-app
```

### 2. Configure as variaveis de ambiente

```bash
cp .env.example .env
```

Edite o `.env` com as credenciais desejadas. O arquivo `.env` nunca e comitado no repositorio — apenas o `.env.example` serve como referencia.

### 3. Gere o go.sum

O arquivo `go.sum` precisa ser gerado uma unica vez antes do primeiro build.
Ele registra os hashes de todas as dependencias e deve ser comitado no repositorio.

```bash
docker run --rm \
  -v $(pwd)/backend:/app \
  -w /app \
  golang:1.26-alpine \
  go mod tidy
```

### 4. Suba o ambiente

```bash
docker compose up --build -d
```

Os containers iniciam na seguinte ordem (controlada pelo `depends_on`):

1. `db` (PostgreSQL) e `redis` — aguardados pelo health check
2. `app` (Go + Air) — sobe apos banco e cache estarem prontos
3. `nginx` — sobe apos o app

### 5. Verifique se esta funcionando

```bash
curl http://localhost:8080/health
```

Resposta esperada:

```json
{
    "status": "ok",
    "timestamp": "2026-04-21T16:00:00Z",
    "services": {
        "postgres": "ok",
        "redis": "ok"
    }
}
```

## Acesso aos servicos

| Servico    | Endereco                     |
| ---------- | ---------------------------- |
| API        | http://localhost:8080        |
| Health     | http://localhost:8080/health |
| PostgreSQL | localhost:5432               |
| Redis      | localhost:6379               |

## Conexao com o banco (DBeaver ou similar)

| Campo    | Valor       |
| -------- | ----------- |
| Host     | localhost   |
| Porta    | 5432        |
| Database | finance_app |
| Usuario  | postgres    |
| Senha    | postgres    |

## Comandos uteis

```bash
# Subir em background
docker compose up -d

# Acompanhar logs em tempo real
docker compose logs -f

# Logs de um servico especifico
docker compose logs -f app

# Encerrar todos os containers
docker compose down

# Encerrar e remover volumes (apaga os dados do banco)
docker compose down -v

# Rebuild forcado apos mudancas no Dockerfile ou dependencias
docker compose up --build -d
```

## Estrutura do projeto

```
finance-app/
├── docker-compose.yml
├── backend/
│   ├── Dockerfile
│   ├── .air.toml               # Configuracao de hot reload
│   ├── go.mod / go.sum
│   └── cmd/api/main.go         # Ponto de entrada e injecao de dependencias
│       internal/
│       ├── database/           # Inicializacao de PostgreSQL e Redis
│       ├── handlers/           # Camada HTTP (equivalente aos Controllers do Laravel)
│       ├── services/           # Logica de negocio (equivalente aos Services do Laravel)
│       ├── repository/         # Acesso a dados (unica camada que fala com o banco)
│       ├── models/             # Entidades de dominio (structs puras)
│       └── utils/              # Utilitarios de dominio
├── frontend/
└── infra/
    └── nginx/nginx.conf        # Proxy reverso com suporte a SSE
```

## Variaveis de ambiente

As variaveis sao definidas no `.env` da raiz e lidas automaticamente pelo Docker Compose.
Copie o `.env.example` para criar o seu:

| Variavel    | Padrao      | Descricao                 |
| ----------- | ----------- | ------------------------- |
| DB_HOST     | db          | Host do PostgreSQL        |
| DB_PORT     | 5432        | Porta do PostgreSQL       |
| DB_USER     | postgres    | Usuario do banco          |
| DB_PASSWORD | postgres    | Senha do banco            |
| DB_NAME     | finance_app | Nome do banco de dados    |
| REDIS_HOST  | redis       | Host do Redis             |
| REDIS_PORT  | 6379        | Porta do Redis            |
| APP_PORT    | 8080        | Porta interna do servidor |
