# Backend — Instruções para o Claude

## Stack

- Go 1.26, Echo v4, GORM, PostgreSQL 17, Redis 7
- Docker Compose para ambiente local (Go não precisa estar instalado localmente)
- Air para hot reload em desenvolvimento
- Nginx como proxy reverso com suporte a SSE (`proxy_buffering off`)
- Migrations: golang-migrate v4 (configurado — adicionar arquivos .sql em backend/migrations/)

## Estilo de comentários no código

Comentários devem ser escritos em pt-br. Identificadores (variáveis, funções,
tipos, campos de struct) devem ser em inglês — é o padrão da linguagem e do
ecossistema Go.

Comentários nos arquivos Go devem sempre incluir a comparação com o equivalente
no Laravel quando o conceito for não-óbvio para quem vem do PHP. Exemplos do
padrão adotado no projeto:

- "Em Laravel, o equivalente é o `__construct()`. Aqui usamos uma função `New*`."
- "Diferente do PHP onde você declara `implements`, em Go qualquer struct que
  implemente os métodos satisfaz a interface automaticamente."
- "Equivalente ao `AppServiceProvider` do Laravel, mas feito de forma explícita
  sem container IoC."

Isso não é opcional: é a convenção pedagógica do projeto. Mantenha esse padrão
ao sugerir novos arquivos ou edições em arquivos existentes.

## Arquitetura

O projeto segue Clean Architecture com três camadas bem definidas:

```
Handler  →  Service  →  Repository
  (HTTP)    (negócio)    (dados)
```

Regras invioláveis:
- Cada camada depende apenas da **interface** da camada abaixo, nunca da
  implementação concreta.
- `main.go` é o único arquivo que conhece as implementações concretas e monta
  o grafo de dependências manualmente.
- Structs de implementação são privadas (nome em minúsculo). O construtor `New*`
  retorna a interface, não o tipo concreto.
- Novas features seguem obrigatoriamente o triplet: arquivo em `handlers/`,
  `services/` e `repository/`, cada um com sua interface definida.

## Decisões técnicas críticas

**Valores monetários:** sempre `int64` representando centavos. Nunca `float64`,
nunca `float32`. Cálculos financeiros com ponto flutuante acumulam erro de
representação binária.

**Injeção de dependências:** manual, sem framework de DI. Toda dependência é
recebida via construtor. Nunca instanciar conexões de banco ou clientes dentro
de um handler ou service.

**Context propagation:** todo método que faz I/O (banco, Redis, HTTP externo)
deve receber `context.Context` como primeiro parâmetro e respeitá-lo. Não use
`context.Background()` dentro de handlers — propague o context da requisição.

**Erros:** retorne `error`, nunca faça panic em código de negócio. Use
`fmt.Errorf("contexto: %w", err)` para empacotar erros com contexto.

## O que não fazer

- Não sugira `AutoMigrate` do GORM para migrations em produção.
- Não use `float64` para dinheiro.
- Não adicione lógica de negócio em handlers.
- Não acesse banco de dados diretamente em services — passe pelo repository.
