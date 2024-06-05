package main

import (
	"database/sql"  // пакет для работы с БД
	"fmt"           // Импорт функции для форматирования строк
	"html/template" // Импорт пакета template для работы с HTML-шаблонами
	"log"           // Импорт пакета для логирования
	"net/http"      // Импорт пакета http для создания HTTP-сервера
	"runtime"
	"sync"
	"time"

	_ "github.com/lib/pq" //Импот драйвера для соединения с сервером PSQL
	"github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache" // Импорт пакета go-cache для кэширования данных
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

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction   string  `json:"transaction"`
	RequestId     string  `json:"request_id"`
	Currency      string  `json:"currency"`
	Provider      string  `json:"provider"`
	Amount        int     `json:"amount"`
	PaymentDt     int     `json:"payment_dt"`
	Bank          string  `json:"bank"`
	DeliveryCost  int     `json:"delivery_cost"`
	GoodsTotal    int     `json:"goods_total"`
	CustomFee     int     `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID         string    `json:"order_uid"`
	TrackNumber      string    `json:"track_number"`
	Entry            string    `json:"entry"`
	Delivery         Delivery  `json:"delivery"`
	Payment          Payment   `json:"payment"`
	Items            []Item    `json:"items"`
	Locale           string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID       string    `json:"customer_id"`
	DeliveryService  string    `json:"delivery_service"`
	ShardKey         string    `json:"shardkey"`
	SmID             int       `json:"sm_id"`
	DateCreated      string    `json:"date_created"`
	OofShard         string    `json:"oof_shard"`
 
}


func New_Cache() *cache.Cache {
	// Создаем кэш с 5 минутами по умолчанию для каждого элемента и чисткой каждые 10 минут
	return cache.New(5*time.Minute, 10*time.Minute)
}


func main() {
 
var cache = New_Cache()
	const DefaultExpiration = 5 * time.Minute	
// Заполняем кэш тестовыми данными
	cache.Set("b563feb7b2b84b6test", 
	Order{OrderUID: "b563feb7b2b84b6test", 
	TrackNumber: "WBILMTESTTRACK", 
	Entry: "WBIL",
	
	}, 
	DefaultExpiration)
	
	
	
	// шаблон HTML
	tmpl, err := template.ParseFiles("templates/order_details.html")
	if err != nil {
		panic(err)
	}



	// Обработчик HTTP-запросов
	fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Получаем ID из запроса
		id := r.URL.Query().Get("id")
		var orderData Order

		if id != "" {
			if value, ok := cache.Get(id); ok {
				orderData = value.(Order)
			}
		}

		// Отображаем шаблон с данными
		tmpl.Execute(w, orderData)
	})

	// Запускаем сервер
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))


	   // Данные для подключения к БД
 const (
	host = "localhost"
	port = 5432
	user = "order_user"
	password = "password"
	dbname = "order_db"
)
	// Подключение к базе данных PostgreSQL
	Psql_info := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname) //инкапсуляция параметров подключения

	order_db, err := sql.Open("postgres", Psql_info)
	if err != nil {
		log.Fatal(err)
	}
	defer order_db.Close()

	// Проверка подключения к базе данных Postgresql
	err = order_db.Ping()
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
	_, err = order_db.Exec(createTable)
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
		_, err := order_db.Exec("INSERT INTO messages (content) VALUES($1)", content)  
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

