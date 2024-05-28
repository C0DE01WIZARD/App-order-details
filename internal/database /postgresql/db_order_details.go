package main

import (
	"database/sql" // пакет для работы с SQL запросами
	"fmt"          // функция для форматирования строк
	"log"
	_ "github.com/lib/pq" // драйвер для соединения с сервером PSQL
)

const (
	host     = "localhost"
	port     = 5432
	user     = "order_user"
	password = "password"
	dbname   = "order_db"
	
)

func main() {
	// Подключение к базе данных PostgreSQL
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

// Создание таблицы в базе данных
	createTable := `
	CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    delivery JSONB,
    
		payment JSONB,
    items JSONB,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard TEXT
);`

// Выполнение запроса на создание таблицы
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Таблица успешно создана!!!")
}
