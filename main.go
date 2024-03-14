package main

import (
	"log"
	"net/http"
	"github.com/vexrina/cinemaLibrary/pkg/database"
	"github.com/vexrina/cinemaLibrary/pkg/userapi"
)



func main(){
	connectionString := "host=172.20.0.2 port=5432 dbname=test_db user=root password=root sslmode=disable"
	db, err := database.ConnectToPG(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) {userapi.RegisterHandler(w,r,db)})
	http.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {userapi.LoginHandler(w,r,db)})
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}