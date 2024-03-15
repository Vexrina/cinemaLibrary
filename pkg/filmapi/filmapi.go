// pkg/filmapi/filmapi.go
package filmapi

import (
	"encoding/json"
	"net/http"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)


func CreateFilmHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var film types.Film
	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert film data to database
	_, err = orm.CreateFilm(film)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateFilmHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var film types.Film
	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = orm.UpdateFilm(film)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}