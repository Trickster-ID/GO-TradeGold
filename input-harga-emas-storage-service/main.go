package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type InputHargaEmas struct {
	Admin_id      string  `json:"admin_id,omitempty"`
	Harga_topup   float32 `json:"harga_topup"`
	Harga_buyback float32 `json:"harga_buyback"`
}

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		panic("fail to load .env")
	}
	kafkaAddress := os.Getenv("kafkaAddress")
	kafkaTopic := os.Getenv("kafkaTopic")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaAddress},
		GroupID:  "jojonomic",
		Topic:    kafkaTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		controller(m)
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}

}

func controller(mes kafka.Message) {
	var input InputHargaEmas

	err := json.Unmarshal([]byte(mes.Value), &input)
	if err != nil {
		log.Fatal("error when unmarshal:", err)
	}
	fmt.Print("struct after : ")
	fmt.Println(input)
	repository(string(mes.Key), input)
}

func repository(reff_id string, input InputHargaEmas) {
	host := os.Getenv("db-host")
	port := os.Getenv("db-port")
	user := os.Getenv("db-user")
	password := os.Getenv("db-pass")
	dbname := os.Getenv("db-name")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatalf("Tidak Konek DB Errornya : %s", err)
	}
	defer db.Close()

	sqlStatement := `INSERT INTO tbl_harga 
	(id, admin_id, harga_topup, harga_buyback, created_date) values
	($1, $2, $3, $4, NOW())
	RETURNING id;`

	var id string
	err = db.QueryRow(sqlStatement, reff_id, input.Admin_id, input.Harga_topup, input.Harga_buyback).Scan(&id)
	if err != nil {
		log.Fatalf("error when execute : %s", err)
	}
	fmt.Println("Success insert with id :", id)

}
