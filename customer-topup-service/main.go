package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/teris-io/shortid"
)

type TopupPayload struct {
	Gram  string `json:"gram"`
	Harga string `json:"harga"`
	Norek string `json:"norek"`
}
type TopupModel struct {
	Gram  float64 `json:"gram"`
	Harga float64 `json:"harga"`
	Norek string  `json:"norek"`
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
	r.HandleFunc("/api/topup", controller).Methods("POST")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	var topupData TopupPayload
	writer.Header().Add("content-type", "application/json")
	json.NewDecoder(request.Body).Decode(&topupData)
	reff_id, err := service(topupData)
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

func service(topupPayload TopupPayload) (string, error) {
	newid, _ := shortid.Generate()
	var topupModel TopupModel
	//try to parse to float
	gramFloat, errConvGram := strconv.ParseFloat(topupPayload.Gram, 64)
	if errConvGram != nil {
		return newid, errConvGram
	}
	topupModel.Gram = gramFloat
	//try to parse to float
	hargaFloat, errConvHarga := strconv.ParseFloat(topupPayload.Harga, 64)
	if errConvHarga != nil {
		return newid, errConvHarga
	}
	topupModel.Harga = hargaFloat
	topupModel.Norek = topupPayload.Norek
	//validate minimal topup
	if gramFloat < 0.001 {
		return newid, errors.New("0.001 is minimal to topup")
	}
	//validate multiple of 0.001
	comma := false
	counter := 0
	for _, v := range topupPayload.Gram {
		sub := fmt.Sprintf("%c", v)
		if comma {
			counter++
		}
		if sub == "." {
			comma = true
		}
	}
	if counter > 3 {
		return newid, errors.New("minimal topup is multiple of 0.001")
	}
	//validate harga topup
	hargaTopup, errVal := validateHarga()
	if errVal != nil {
		return newid, errVal
	}
	if topupModel.Harga != hargaTopup {
		return newid, errors.New("harga is not match by topup harga")
	}

	err := repository(newid, topupModel)
	if err != nil {
		return newid, err
	}
	return newid, nil
}

func repository(reff_id string, topupModel TopupModel) error {
	kafkaAddress := os.Getenv("kafkaAddress")
	kafkaTopic := os.Getenv("kafkaTopic")

	w := &kafka.Writer{
		Addr:     kafka.TCP(kafkaAddress),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	res, _ := json.Marshal(topupModel)
	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(reff_id),
			Value: res,
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
		return err
	}

	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
		return err
	}
	fmt.Println(kafkaTopic)
	return nil
}

func validateHarga() (float64, error) {
	host := os.Getenv("db-host")
	port := os.Getenv("db-port")
	user := os.Getenv("db-user")
	password := os.Getenv("db-pass")
	dbname := os.Getenv("db-name")
	var hargaTopup float64

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return hargaTopup, err
	}
	defer db.Close()

	sqlStatement := `SELECT harga_topup FROM tbl_harga ORDER BY created_date DESC LIMIT 1;`
	errExec := db.QueryRow(sqlStatement).Scan(&hargaTopup)
	if errExec != nil {
		return hargaTopup, errExec
	}
	return hargaTopup, nil
}
