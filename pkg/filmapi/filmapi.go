// pkg/filmapi/filmapi.go
package filmapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

// post method
func CreateFilmHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var film types.Film
	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil || film.Title == "" {
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

// patch method
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

// get method
// utility
type EnumType string

const (
	EnumValue1 EnumType = "rating"
	EnumValue2 EnumType = "title"
	EnumValue3 EnumType = "release_date"
)

func IsValidEnumType(value EnumType) bool {
	switch value {
	case EnumValue1, EnumValue2, EnumValue3:
		return true
	default:
		return false
	}
}

func ValidateEnumType(value EnumType) error {
	if !IsValidEnumType(value) {
		return errors.New("invalid enum value")
	}
	return nil
}

func ReturnAnswer(response []types.Film, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// main func for get method
func GetFilmsHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	queryValues := r.URL.Query()

	if len(queryValues) == 0 {
		films, err := orm.GetFilms("", false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}
	// validate sortBy
	sortBy, sortOk := queryValues["sortby"]
	if sortOk {
		err := ValidateEnumType(EnumType(sortBy[0]))
		if err != nil {
			http.Error(w, "Invalid request for sortBy", http.StatusBadRequest)
		}
	}

	// validate Ascendion
	ascStr, ascOk := queryValues["asc"]
	asc := false
	if ascOk {
		ascBool, err := strconv.ParseBool(ascStr[0])
		if err != nil {
			http.Error(w, "Invalid value for ascending parameter", http.StatusBadRequest)
		}
		asc = ascBool
	}
	if sortOk {
		if ascOk {
			films, err := orm.GetFilms(sortBy[0], asc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			ReturnAnswer(films, w, r)
			return
		}
		films, err := orm.GetFilms(sortBy[0], false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}

	if ascOk {
		films, err := orm.GetFilms("", asc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}

	actor, okActor := queryValues["actor"]
	
	title, okTitle := queryValues["title"]
	
	actorTitle, okBoth := queryValues["actor_title"]
	
	if (okActor && okTitle) || (okTitle && okBoth) || (okBoth && okActor) {
		http.Error(w, "Invalid request for filtеr", http.StatusBadRequest)
		return
	}
	if okActor {
		films, err := orm.SearchFilmsByActorFragment(actor[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}
	if okTitle {
		films, err := orm.SearchFilmsByTitleFragment(title[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}
	if okBoth {
		films, err := orm.SearchFilmsByFragment(actorTitle[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		ReturnAnswer(films, w, r)
		return
	}

	http.Error(w, "Invalid URL format", http.StatusBadRequest)
}

// delete method
func DeleteFilmHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var film types.Film
	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if film.ID == 0 {
		http.Error(w, "Film ID is required", http.StatusBadRequest)
		return
	}
	
	err = orm.DeleteFilmByID(film.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
