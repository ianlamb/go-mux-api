package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type GraphqlBody struct {
	query interface{}
}

const (
	HOST = "db"
	PORT = 5432
)

func (a *App) Initialize(user, password, dbname string) {
	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%v/%s?sslmode=disable", user, password, HOST, PORT, dbname)
	log.Printf("Connecting to DB with connection string: %s", connectionString)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Println("Listening on port 8010")
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/items", a.getItems).Methods("GET")
	a.Router.HandleFunc("/item", a.createItem).Methods("POST")
	a.Router.HandleFunc("/item/{id:[0-9]+}", a.getItem).Methods("GET")
	a.Router.HandleFunc("/item/{id:[0-9]+}", a.updateItem).Methods("PUT")
	a.Router.HandleFunc("/item/{id:[0-9]+}", a.deleteItem).Methods("DELETE")

	// graphql
	a.Router.HandleFunc("/graphql/item", func(w http.ResponseWriter, r *http.Request) {
		var query string

		log.Println("Content-Type: ", r.Header.Get("Content-Type"))
		if r.Header.Get("Content-Type") == "application/json" {
			var bodyJson map[string]interface{}
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&bodyJson); err != nil {
				respondWithError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			defer r.Body.Close()
			log.Println("bodyJson: ", bodyJson)

			query = bodyJson["query"].(string)
		} else {
			query = r.URL.Query().Get("query")
		}
		log.Println("Query: ", query)

		result := executeQuery(query, schema, a.DB)
		json.NewEncoder(w).Encode(result)
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *App) getItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	i := item{ID: id}
	if err := i.getItem(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Item not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, i)
}

func (a *App) getItems(w http.ResponseWriter, r *http.Request) {
	items, err := getItems(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	var i item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := i.createItem(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, i)
}

func (a *App) updateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var i item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	i.ID = id

	if err := i.updateItem(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, i)
}

func (a *App) deleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	i := item{ID: id}
	if err := i.deleteItem(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, i)
}
