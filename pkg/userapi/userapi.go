package userapi

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	
	"github.com/vexrina/cinemaLibrary/pkg/types"
)


func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(w, err)
		return
	}

	// check that such an email-username pair is unique
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1 AND email=$2", user.Username, user.Email).Scan(&count)
	if err != nil || count>0 {
		// log.Println(w, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", user.Username, user.Email, string(hashedPassword))
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get hashed password from db
	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE email=$1", user.Email).Scan(&storedPassword)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}