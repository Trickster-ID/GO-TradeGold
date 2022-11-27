package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type RequestModel struct {
	Norek      string `json:"norek"`
	Start_date string `json:"start_date"`
	End_date   string `json:"end_date"`
}
type ResponseModel struct {
	Date          time.Time `json:"date"`
	Type          string    `json:"type"`
	Gram          float64   `json:"gram"`
	Harga_topup   float64   `json:"harga_topup"`
	Harga_buyback float64   `json:"harga_buyback"`
	Saldo         float64   `json:"saldo"`
}
type Success struct {
	Error bool            `json:"error"`
	Data  []ResponseModel `json:"data"`
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
	r.HandleFunc("/api/mutasi", controller).Methods("GET")
	fmt.Println("application is listening...")
	http.ListenAndServe(":"+os.Getenv("port"), r)
}

func controller(writer http.ResponseWriter, request *http.Request) {
	var reqMod RequestModel
	writer.Header().Add("content-type", "application/json")
	json.NewDecoder(request.Body).Decode(&reqMod)
	res, err := repository(reqMod)
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

func repository(reqMod RequestModel) ([]ResponseModel, error) {
	var saldo []ResponseModel
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
	sqlStatement :=
		`SELECT date, type, gram, harga_topup, harga_buyback, saldo FROM (
		SELECT created_date as date, 'topup' as type, gram, harga_topup, harga_buyback, norek, saldo  FROM tbl_topup a
		UNION
		SELECT created_date as date, 'buyback' as type, gram, harga_topup, harga_buyback, norek, saldo  FROM tbl_transaksi c
	) mutasi
	WHERE norek = $1 and (date >= $2 and date <= $3 )`
	rows, errExec := db.Query(sqlStatement, reqMod.Norek, reqMod.Start_date, reqMod.End_date)
	if errExec != nil {
		//log.Fatalf("error when execute : %s", errExec)
		return saldo, errExec
	}
	for rows.Next() {
		var r ResponseModel
		errScan := rows.Scan(&r.Date, &r.Type, &r.Gram, &r.Harga_topup, &r.Harga_buyback, &r.Saldo)
		if errScan != nil {
			return saldo, errScan
		}
		saldo = append(saldo, r)
	}
	return saldo, nil
}
