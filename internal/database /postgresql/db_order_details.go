package main

import (
				_"github.com/lib/pq" // драйвер для соединения с сервером PSQL
				"database/sql" // пакет для работы с SQL запросами
				"fmt" // функция для форматирования строк
				"log"
	 			
)

const (
			host = "localhost"
			port = 5432
			user = "Order_user"
			password = "password"
			dbname = "order_db"

)
func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Table created successfully.")
}
