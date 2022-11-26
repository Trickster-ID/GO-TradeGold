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

type RequestModel struct {
	Norek string `json:"norek"`
}
type ResponseModel struct {
	Norek string  `json:"norek"`
	Saldo float64 `json:"saldo"`
}
type Success struct {
	Error bool          `json:"error"`
	Data  ResponseModel `json:"data"`
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
	r.HandleFunc("/api/saldo", controller).Methods("GET")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	var norek RequestModel
	writer.Header().Add("content-type", "application/json")
	json.NewDecoder(request.Body).Decode(&norek)
	res, err := repository(norek.Norek)
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

func repository(norek string) (ResponseModel, error) {
	var saldo ResponseModel
	host := os.Getenv("db-host")
	port := os.Getenv("db-port")
	user := os.Getenv("db-user")
	password := os.Getenv("db-pass")
	dbname := os.Getenv("db-name")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		//log.Fatalf("Tidak Konek DB Errornya : %s", err)
		return saldo, err
	}
	defer db.Close()
	sqlStatement := "SELECT norek, saldo FROM tbl_rekening WHERE norek = $1 LIMIT 1;"
	errExec := db.QueryRow(sqlStatement, norek).Scan(&saldo.Norek, &saldo.Saldo)
	if errExec != nil {
		//log.Fatalf("error when execute : %s", errExec)
		return saldo, errExec
	}
	return saldo, nil
}
