package main

import (
	"html/template"
	"database/sql"
	"fmt" // функция для форматирования строк
	"log"
	"runtime"


	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // драйвер для соединения с сервером PSQL
	"github.com/nats-io/stan.go"
)


func main() {
    // Загрузка переменных окружения из файла .env
	err := godotenv.Load("/home/rushan/App_order_details/cmd/.env")
	if err != nil {
		fmt.Println("Ошибка при загрузке файла .env:", err)
		return
	}

    // Подключение к базе данных
 const (
	host     = "localhost"
	port     = 5432
	user     = "order_user"
	password = "password"
	dbname   = "order_db"
	
)

	// Подключение к базе данных PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных:", err)
	}

	// Создание таблицы в базе данных
createTable := `
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(50),
    locale VARCHAR(5),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey INT,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);


CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR(255),
    name VARCHAR(100),
    phone VARCHAR(50),
    zip VARCHAR(20),
    city VARCHAR(100),
    address VARCHAR(255),
    region VARCHAR(100),
    email VARCHAR(100),
    PRIMARY KEY (order_uid),
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);


CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR(255),
    transaction UUID,
    request_id VARCHAR(255),
    currency VARCHAR(5),
    provider VARCHAR(50),
    amount INT,
    payment_dt TIMESTAMP,
    bank VARCHAR(50),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT,
    PRIMARY KEY (order_uid),
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE TABLE IF NOT EXISTS items (
    item_id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255),
    track_number VARCHAR(50),
    price INT,
    rid UUID,
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id INT,
    brand VARCHAR(100),
    status INT,
    FOREIGN KEY (order_uid) REFERENCES orders (order_uid)
);
`


// Выполнение запроса на создание таблицы
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Таблица успешно создана!!!")


	// Установление соединение с сервером NATS Streaming
	clusterID := "cluster"
	clientID := "client"
	natsURL := "nats://0.0.0.0:4222"

    // Подключение к серверу NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Не удалось подключиться к NATS Streaming: %v", err)
	}
	defer sc.Close()



	// Подписка на канал
	subject := "channel"
	sub, err := sc.Subscribe(subject, func(m *stan.Msg) {
		log.Printf("Получено сообщение: %s\n", string(m.Data))
	}, stan.DeliverAllAvailable())
	if err != nil {
		log.Fatalf("Не удалось подписаться на канал: %v", err)
	}
	defer sub.Unsubscribe()
	
// Ждем сигнала завершения
	runtime.Goexit()
}