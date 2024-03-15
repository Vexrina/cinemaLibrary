package main

import (
	"log"
	"net/http"

	"github.com/vexrina/cinemaLibrary/pkg/actorapi"
	"github.com/vexrina/cinemaLibrary/pkg/database"
	"github.com/vexrina/cinemaLibrary/pkg/filmapi"
	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/tokens"
	"github.com/vexrina/cinemaLibrary/pkg/userapi"
)

func main() {
	connectionString := "host=172.20.0.3 port=5432 dbname=test_db user=root password=root sslmode=disable"
	db, err := database.ConnectToPG(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	actorOrm := orm.NewORM(db)
	filmOrm := orm.NewORM(db)
	userOrm := orm.NewORM(db)

	http.HandleFunc("/actor", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			admin, err := tokens.ValidateToken(w, r) // maybe it's worth rescheduling this to switch? not sure.
			if err != nil || !admin {                // notNil || admin = 0 -> error
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				actorapi.CreateActorHandler(w, r, actorOrm)
			}
		case http.MethodPatch:
			admin, err := tokens.ValidateToken(w, r)
			if err != nil || !admin { // nil || !admin = 0 -> error
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				actorapi.UpdateActorHandler(w, r, actorOrm)
			}
		case http.MethodDelete:
			admin, err := tokens.ValidateToken(w, r)
			if err != nil || !admin { // notNil || !admin = 0 -> error
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				actorapi.DeleteActorHandler(w, r, actorOrm)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/film", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			admin, err := tokens.ValidateToken(w, r)
			if err != nil || !admin {
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				filmapi.CreateFilmHandler(w, r, filmOrm)
			}
		case http.MethodPatch:
			admin, err := tokens.ValidateToken(w, r)
			if err != nil || !admin {
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				filmapi.UpdateFilmHandler(w, r, filmOrm)
			}
		case http.MethodGet:
			_, err := tokens.ValidateToken(w, r)
			if err != nil {
				http.Error(w, "Bad token", http.StatusUnauthorized)
			} else {
				filmapi.GetFilmsHandler(w, r, filmOrm)
			}
		case http.MethodDelete:
			admin, err := tokens.ValidateToken(w, r)
			if err != nil || !admin {
				http.Error(w, "Method not allowed by role", http.StatusMethodNotAllowed)
			} else {
				filmapi.DeleteFilmHandler(w, r, filmOrm)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) { userapi.RegisterHandler(w, r, userOrm) })
	http.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) { userapi.LoginHandler(w, r, userOrm) })

	log.Fatal(http.ListenAndServe(":8080", nil))
}
