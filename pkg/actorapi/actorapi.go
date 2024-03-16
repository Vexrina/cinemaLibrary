// pkg/actorapi/actorapi.go
package actorapi

import (
	"encoding/json"
	"net/http"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

// method post
func CreateActorHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var actor types.Actor
	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = orm.CreateActor(actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// method patch
func UpdateActorHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
	var actor types.Actor
	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = orm.UpdateActor(actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// method delete
func DeleteActorHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
    var actor types.Actor
	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if actor.ID ==0 {
		http.Error(w, "Actor ID is required", http.StatusBadRequest)
		return
	}

	err = orm.DeleteActorByID(actor.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

    // Возвращаем успешный статус
    w.WriteHeader(http.StatusOK)
}

// method get
func GetActorsHandler(w http.ResponseWriter, r *http.Request, orm *orm.ORM) {
    // url like /actors?fragment={fragment}
    fragment := r.URL.Query().Get("fragment")
    if fragment != "" {
		// find by fragment
        actors, err := orm.GetActorsWithFragment(fragment)
        if err != nil {
            http.Error(w, "database error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(actors)
    } else {
        // find all actors
        actors, err := orm.GetActors()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(actors)
    }
}