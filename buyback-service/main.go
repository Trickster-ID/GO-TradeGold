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

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/teris-io/shortid"
)

type BuybackModel struct {
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
	r.HandleFunc("/api/buyback", controller).Methods("POST")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	var buybackData BuybackModel
	writer.Header().Add("content-type", "application/json")
	json.NewDecoder(request.Body).Decode(&buybackData)
	reff_id, err := service(buybackData)
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

func service(buyback BuybackModel) (string, error) {
	newid, _ := shortid.Generate()
	//validate norek and saldo
	saldo, errValNorek := validateNorek(buyback.Norek)
	if errValNorek != nil {
		return newid, errors.New("error Validate Norek : " + errValNorek.Error())
	}
	if saldo < buyback.Gram {
		return newid, errors.New("buyback value more than saldo")
	}

	//validate minimal buyback
	if buyback.Gram < 0.001 {
		return newid, errors.New("0.001 is minimal to buyback")
	}
	//validate multiple of 0.001
	comma := false
	counter := 0
	for _, v := range fmt.Sprintf("%v", buyback.Gram) {
		sub := fmt.Sprintf("%c", v)
		if comma {
			counter++
		}
		if sub == "." {
			comma = true
		}
	}
	if comma && counter > 3 {
		return newid, errors.New("minimal buyback is multiple of 0.001")
	}
	//validate harga buyback
	hargaBuyback, errVal := validateHarga()
	if errVal != nil {
		return newid, errVal
	}
	if buyback.Harga != hargaBuyback {
		return newid, errors.New("harga is not match by buyback harga")
	}

	err := repository(newid, buyback)
	if err != nil {
		return newid, err
	}
	return newid, nil
}

func repository(reff_id string, buybackModel BuybackModel) error {
	kafkaAddress := os.Getenv("kafkaAddress")
	kafkaTopic := os.Getenv("kafkaTopic")

	w := &kafka.Writer{
		Addr:     kafka.TCP(kafkaAddress),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	res, _ := json.Marshal(buybackModel)
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
	db, err := dbConfig()
	var hargaBuyback float64
	if err != nil {
		return hargaBuyback, err
	}
	defer db.Close()

	sqlStatement := `SELECT harga_buyback FROM tbl_harga ORDER BY created_date DESC LIMIT 1;`
	errExec := db.QueryRow(sqlStatement).Scan(&hargaBuyback)
	if errExec != nil {
		return hargaBuyback, errExec
	}
	return hargaBuyback, nil
}

func validateNorek(norek string) (float64, error) {
	db, errDb := dbConfig()
	if errDb != nil {
		return 0, errDb
	}
	defer db.Close()
	sqlStatement := `SELECT saldo FROM tbl_rekening WHERE norek = $1;`
	var saldo float64
	errExec := db.QueryRow(sqlStatement, norek).Scan(&saldo)
	if errExec != nil {
		return 0, errors.New("norek not exist")
	}
	return saldo, nil
}

func dbConfig() (*sql.DB, error) {
	host := os.Getenv("db-host")
	port := os.Getenv("db-port")
	user := os.Getenv("db-user")
	password := os.Getenv("db-pass")
	dbname := os.Getenv("db-name")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	return sql.Open("postgres", psqlInfo)
}
