package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

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

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	orderJSON := `{"order_uid": "b563feb7b2b84b6test", "track_number": "WBILMTESTTRACK", ...}`

	var order Order
	err = json.Unmarshal([]byte(orderJSON), &order)
	if err != nil {
		panic(err)
	}

	// Сериализация структуры заказа в JSON
	serializedOrder, err := json.Marshal(order)
	if err != nil {
		panic(err)
	}

	// Сохранение сериализованного заказа в базу данных
	_, err = db.Exec("INSERT INTO orders(data) VALUES($1)", serializedOrder)
	if err != nil {
		panic(err)
	}

	fmt.Println("Заказ сохранен в базе данных.")
}