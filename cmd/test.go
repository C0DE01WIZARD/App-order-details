package main
import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)
 
func () { 
 
    connStr := "user=postgres password=mypass dbname=productdb sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err)
    } 
    defer db.Close()
     
    result, err := db.Exec("insert into Products (model, company, price) values ('iPhone X', $1, $2)", 
        "Apple", 72000)
    if err != nil{
        panic(err)
    }
    fmt.Println(result.LastInsertId())  // не поддерживается
    fmt.Println(result.RowsAffected())  // количество добавленных строк
}