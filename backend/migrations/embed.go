// Package migrations expõe os arquivos SQL de migration para serem embedados
// no binário de produção via diretiva //go:embed do Go.
//
// O embedding garante que o binário gerado no stage "production" do Dockerfile
// já carrega todas as migrations dentro de si — sem precisar montar volumes ou
// copiar arquivos SQL para o container. É o equivalente a ter as migrations
// versionadas junto com o código, não como artefato externo.
//
// Em Laravel, as migrations ficam em database/migrations/ e são lidas do
// sistema de arquivos em runtime. Aqui, os arquivos SQL são compilados dentro
// do próprio binário Go — mais seguro e portável em ambientes containerizados.
package migrations

import "embed"

// FS é o sistema de arquivos embedado que contém todas as migrations SQL.
//
// Para ativar: quando o primeiro arquivo .sql for criado neste diretório,
// adicione a diretiva abaixo imediatamente antes do "var FS":
//
//	//go:embed *.sql
//
// A partir daí, qualquer arquivo *.sql adicionado aqui será automaticamente
// incluído no binário na próxima compilação.
var FS embed.FS
