// Package models contem as entidades de dominio da aplicacao.
//
// As entidades sao structs puras que representam os conceitos centrais do negocio.
// Elas nao contem logica — apenas dados e suas anotacoes de mapeamento (GORM, JSON).
// Toda logica que opera sobre uma entidade vive nos services.
//
// Em Laravel, o equivalente sao os Models do Eloquent (ex: app/Models/User.php),
// com a diferenca que la o Model mistura definicao de dados com acesso ao banco.
// Em Go, separamos essas responsabilidades: a struct define os dados (aqui),
// e o repositorio define como persisti-los (internal/repository/).
package models

import (
	"time"
)

// User representa um usuario cadastrado na plataforma de planejamento financeiro.
// Os campos exportados (iniciados em maiusculo) sao acessiveis fora do pacote.
// As tags `gorm:"..."` instruem o ORM sobre restricoes de banco de dados.
// As tags `json:"..."` controlam a serializacao para respostas da API.
type User struct {
	// ID e a chave primaria auto-incrementada. GORM reconhece o nome "ID"
	// como chave primaria por convencao, sem necessidade de anotacao adicional.
	ID uint `gorm:"primaryKey" json:"id"`

	// Name e o nome completo do usuario. A restricao not null e aplicada
	// diretamente no schema do banco via GORM.
	Name string `gorm:"not null" json:"name"`

	// Email e o endereco de email do usuario, unico na base de dados.
	// O uniqueIndex cria um indice de unicidade no PostgreSQL, garantindo
	// consistencia mesmo em insercoes concorrentes — algo que validacao
	// em nivel de aplicacao nao consegue garantir sozinha.
	Email string `gorm:"uniqueIndex;not null" json:"email"`

	// CreatedAt e UpdatedAt sao preenchidos automaticamente pelo GORM
	// nas operacoes de criacao e atualizacao, respectivamente.
	// Equivalente aos timestamps() do Laravel nas migrations.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
