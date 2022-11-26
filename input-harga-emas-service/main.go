package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/teris-io/shortid"
)

type InputHargaEmas struct {
	Admin_id      string  `json:"admin_id,omitempty"`
	Harga_topup   float64 `json:"harga_topup"`
	Harga_buyback float64 `json:"harga_buyback"`
}
type Success struct {
	Error   bool   `json:"error"`
	Reff_id string `json:"reff_id"`
}
type Error struct {
	Error   bool   `json:"error"`
	Reff_id string `json:"reff_id"`
	Message string `json:"message"`
}

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		panic("fail to load .env")
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/input-harga", controller).Methods("POST")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	var input InputHargaEmas
	writer.Header().Add("content-type", "application/json")
	json.NewDecoder(request.Body).Decode(&input)
	reff_id, err := repository(input)
	if err != nil {
		res := Error{
			Error:   true,
			Reff_id: reff_id,
			Message: err.Error(),
		}
		writer.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(writer).Encode(res)
		return
	}
	res := Success{
		Error:   false,
		Reff_id: reff_id,
	}
	json.NewEncoder(writer).Encode(res)
}

func repository(input InputHargaEmas) (string, error) {
	newid, _ := shortid.Generate()
	kafkaAddress := os.Getenv("kafkaAddress")
	kafkaTopic := os.Getenv("kafkaTopic")

	w := &kafka.Writer{
		Addr:     kafka.TCP(kafkaAddress),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	res, _ := json.Marshal(input)
	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(newid),
			Value: res,
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
		return newid, err
	}

	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
		return newid, err
	}
	fmt.Println(kafkaTopic)
	return newid, nil
}
