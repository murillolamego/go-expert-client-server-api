package main

import "database/sql"

func main() {

	const create string = `
	CREATE TABLE IF NOT EXISTS usdbrl (
	id INTEGER NOT NULL PRIMARY KEY,
	code TEXT,
	codein TEXT,
	name TEXT,
	high TEXT,
	low TEXT,
	var_bid TEXT,
	pct_change TEXT,
	bid TEXT,
	ask TEXT,
	timestamp TEXT,
	create_date TEXT
	);`

	db, err := sql.Open("sqlite3", "usdbrl.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(create)
	if err != nil {
		panic(err)
	}

	Server()
}
