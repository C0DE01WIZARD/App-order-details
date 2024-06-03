package main

import (
	"html/template"
	"database/sql" // пакет для работы с БД 
	"sync"
	"net/http"
	_"encoding/json"
	"fmt" // функция для форматирования строк
	"log" // 
	"runtime"
	_ "github.com/lib/pq" // драйвер для соединения с сервером PSQL
	"github.com/nats-io/stan.go"
)


/* Функция для восстановления кэша из БД, вызывается в начале функции main
для того чтобы кэш восстаавливался перед началом обработки запросов*/
func (c *Cache) RestoreFromDB(db *sql.DB) error {
    // Запрос для получения всех данных из таблицы orders
    rows, err := db.Query("SELECT order_uid, track_number FROM orders")
    if err != nil {
        return err
    }
    defer rows.Close()

    // Чтение результатов запроса и восстановление кэша
    for rows.Next() {
        var orderUID, trackNumber string
        if err := rows.Scan(&orderUID, &trackNumber); err != nil {
            return err
        }
        c.Set(orderUID, trackNumber)
    }

    // Проверка на ошибки при получении данных
    if err = rows.Err(); err != nil {
        return err
    }

    fmt.Println("Кэш восстановлен из БД")
    return nil
}


func main() {
    // Данные для подключения к БД
 const (
	host = "localhost"
	port = 5432
	user = "order_user"
	password = "password"
	dbname = "order_db"
)

template_html, err := template.ParseFiles("templates/order_details.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
            http.Error(w, "Не указан ID", http.StatusBadRequest)
            return
        }
				
		template_html.Execute(w, nil)
	})

	http.ListenAndServe(":8080", nil)
	
	// Подключение к базе данных PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка подключения к базе данных Postgresql
	err = db.Ping()
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных:", err)
	}

	
	// 1.1 Создание базы данных Postgresql
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

CREATE TABLE IF NOT EXISTS messages(
	  id SERIAL PRIMARY KEY,
		content TEXT NOT NULL
);
`

// Выполнение запроса на создание таблицы
	_, err = db.Exec(createTable)
	if err != nil { // обработка ошибок БД
		log.Fatal(err)
	}
	fmt.Println("Таблица dbname успешно создана и работает на порту:", port)


	// Данные для соединение с сервером NATS Streaming
	clusterID := "my_cluster"
	clientID := "my_client_id"
	natsURL := "nats://0.0.0.0:4222"

    // Подключение к серверу NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil { // обработка ошибки подключения NS
		log.Fatalf("Не удалось подключиться к NATS Streaming: %v", err)
	}
	defer sc.Close()
	fmt.Println("Успешное подключение к адресу:"+natsURL)

	// Подписка на канал
	subject := "Order_details_channel"
	sub, err := sc.Subscribe(subject, func(m *stan.Msg){
		log.Printf("Получено сообщение: %s\n", string(m.Data))
		
		// Запись данных в БД
	content := "Запись в БД"
		_, err := db.Exec("INSERT INTO messages (content) VALUES($1)", content)  
		if err != nil { // обработка ошибки записи в БД
			log.Fatalf("Ошибка при записи в базу данных: %v", err)
		}
	}, stan.DeliverAllAvailable())
	if err != nil { 
		log.Fatalf("Не удалось подписаться на канал: %v", err)
	}
	defer sub.Unsubscribe()
	fmt.Println("Подписка на канал", subject, "и запись в БД успешно выполнена.")
	
// Ожидание сигнала завершения
	runtime.Goexit()
}

/*2.3 Реализовать кэширование полученных данных в сервисе (сохранять in memory)
 Структура для хранения кэша */

type Cache struct {
	sync.RWMutex
	store map[string]string
}

// NewCache создает новый экземпляр Cache
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]string),
	}
}

// Set добавляет пару ключ-значение в кэш
func (c *Cache) Set(key string, value string) {
	c.Lock()
	defer c.Unlock()
	c.store[key] = value
}

// Get возвращает значение по ключу из кэша
func(c *Cache) Get(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

func Create_cash() {
	// Создание кэша
	cache := NewCache()

	// Добавление данных в кэш
	cache.Set("ключ", "значение")

	// Получение данных из кэша
	if val, ok := cache.Get("ключ"); ok {
		println("Найдено в кэше:", val)
		} else {
		println("Значение не найдено в кэше.")
	}
}
