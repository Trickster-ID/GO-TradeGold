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

type BuybackModel struct {
	Gram  float64 `json:"gram"`
	Harga float64 `json:"harga"`
	Norek string  `json:"norek"`
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
	fmt.Println("application is listening...")
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
	var buybackData BuybackModel

	err := json.Unmarshal([]byte(mes.Value), &buybackData)
	if err != nil {
		log.Fatal("error when unmarshal:", err)
	}
	fmt.Print("struct after : ")
	fmt.Println(buybackData)
	fmt.Println(buybackData.Gram)
	repository(string(mes.Key), buybackData)
}

func repository(reff_id string, input BuybackModel) {
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

	sqlStatement := `CALL sp_buyback($1,$2,$3)`
	// _, errExec := db.Query(sqlStatement, "1234qwer", 0.2, 800000, "234r")
	// _, errExec := db.Query(sqlStatement, reff_id, input.Gram, input.Harga, input.Norek)
	_, errExec := db.Exec(sqlStatement, reff_id, input.Gram, input.Norek)
	if errExec != nil {
		log.Fatalf("error when execute : %s", errExec)
	}
	fmt.Println("Success insert with id :", reff_id)
}
