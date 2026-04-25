# Finance App — Instruções para o Claude

## Contexto do projeto

SaaS de planejamento financeiro. O time é formado por dois desenvolvedores
aprendendo juntos, ambos com background em PHP/Laravel. O objetivo é aprender
com profundidade, não apenas fazer o código funcionar.

Toda sugestão e explicação deve levar isso em conta: a audiência conhece bem
Laravel e está construindo o equivalente mental nas tecnologias usadas aqui.

## Estrutura do repositório

```
finance-app/
├── backend/    ← API em Go (ver backend/CLAUDE.md para regras específicas)
├── frontend/   ← a definir
└── infra/      ← Nginx, Docker Compose
```

## Regras transversais

- Comentários no código e mensagens de commit devem ser escritos em pt-br.
  Identificadores (variáveis, funções, tipos, constantes) devem ser em inglês.
  Essa regra vale em qualquer camada — backend, frontend.
- Não crie abstrações ou helpers além do que a tarefa imediata exige.
- Não adicione features, refatorações ou cleanup além do que foi pedido.
- Valores monetários são sempre representados em centavos (`int64`). Essa regra
  vale em qualquer camada — backend, frontend, banco de dados.
