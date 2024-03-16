package userapi

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/tokens"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)


func RegisterHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if user.Email=="" || user.Password=="" || user.Username ==""{
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	count, err := orm.CountUsersWithUsernameAndEmail(user.Username, user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Username or email already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = orm.CreateUser(user.Username, user.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storedPassword, adminflag, err := orm.GetUserPasswordByEmail(user.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := tokens.CreateToken(user.Username, adminflag)
	if err != nil {
		http.Error(w, "Error with creating token", http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	response := map[string]string{"token":tokenString}
	
	err = json.NewEncoder(w).Encode(response)
	if err!=nil{
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}