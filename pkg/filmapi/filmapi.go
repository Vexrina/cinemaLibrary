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

type EnumType string

const (
	enumValue1 EnumType = "rating"
	enumValue2 EnumType = "title"
	enumValue3 EnumType = "release_date"
)

func IsValidEnumType(value EnumType) bool {
	switch value {
	case enumValue1, enumValue2, enumValue3:
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

func GetFilmsHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	queryValues := r.URL.Query()

	if len(queryValues) == 0 {
		orm.GetFilms("", false)
		return
	}

	// validate sortBy
	sortBy, sortOk := queryValues["sortby"]
	err := ValidateEnumType(EnumType(sortBy[0]))
	if err != nil {
		http.Error(w, "Invalid request for sortBy", http.StatusBadRequest)
		return
	}

	// validate Ascendion
	ascStr, ascOk := queryValues["asc"]
	asc, err := strconv.ParseBool(ascStr[0])
	if err != nil {
		http.Error(w, "Invalid value for ascending parameter", http.StatusBadRequest)
		return
	}

	if sortOk {
		if ascOk {
			orm.GetFilms(sortBy[0], asc)
			return
		}
		orm.GetFilms(sortBy[0], false)
	}

	if ascOk {
		orm.GetFilms("", asc)
	}

	actor, okActor := queryValues["actor"]

	title, okTitle := queryValues["title"]

	actorTitle, okBoth := queryValues["actor_title"]

	if (okActor && okTitle) || (okTitle && okBoth) || (okBoth && okActor) {
		http.Error(w, "Invalid request for filtr", http.StatusBadRequest)
		return
	}

	if okActor {
		orm.SearchFilmsByActorFragment(actor[0])
		return
	}
	if okTitle {
		orm.SearchFilmsByTitleFragment(title[0])
		return
	}
	if okBoth {
		orm.SearchFilmsByFragment(actorTitle[0])
	}

	http.Error(w, "Invalid URL format", http.StatusBadRequest)
}
