package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type HargaEmas struct {
	Harga_topup   float32 `json:"harga_topup"`
	Harga_buyback float32 `json:"harga_buyback"`
}
type Success struct {
	Error bool      `json:"error"`
	Data  HargaEmas `json:"data"`
}
type Error struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		panic("fail to load .env")
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/check-harga", controller).Methods("GET")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("content-type", "application/json")
	res, err := repository()
	if err != nil {
		res := Error{
			Error:   true,
			Message: err.Error(),
		}
		writer.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(writer).Encode(res)
		return
	}
	response := Success{
		Error: false,
		Data:  res,
	}
	json.NewEncoder(writer).Encode(response)
}

func repository() (HargaEmas, error) {
	var hargaEmas HargaEmas
	host := os.Getenv("db-host")
	port := os.Getenv("db-port")
	user := os.Getenv("db-user")
	password := os.Getenv("db-pass")
	dbname := os.Getenv("db-name")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return hargaEmas, err
	}
	defer db.Close()
	sqlStatement := "SELECT harga_topup, harga_buyback FROM tbl_harga ORDER BY created_date DESC LIMIT 1;"
	errExec := db.QueryRow(sqlStatement).Scan(&hargaEmas.Harga_topup, &hargaEmas.Harga_buyback)
	if errExec != nil {
		return hargaEmas, errExec
	}
	return hargaEmas, nil
}
