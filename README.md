# Client-Server-API

Primeiro desafio da Pós Graduação - Go Expert da Full Cycle

- client.go

Faz requisição ao server.go, recebe uma cotação no formato { "bid": "" } e deverá grava-lá com o formato "Dólar: { valor }" no arquivo "cotacao.txt".

Limite de tempo padrão de execução: 300ms.

- server.go

Escuta na porta 8080, na rota /cotacao deverá consumir a API https://economia.awesomeapi.com.br/json/last/USD-BRL e retornar um JSON com formato { "bid": "" }, além de persistir a cotação na tabela "usdbrl.db" (que será criada no SQLite caso não exista).

Limite de tempo padrão da requisição à API: 200ms.

Limite de tempo padrão para persistir no banco de dados: 10ms.
