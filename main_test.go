package main_test

import (
	"bytes"
	"encoding/json"
	"github.com/ianlamb/go-mux-api"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM items")
	a.DB.Exec("ALTER SEQUENCE items_id_seq RESTART WITH 1")
}

func addItems(qty int) {
	if qty < 1 {
		qty = 1
	}

	for i := 0; i < qty; i++ {
		itemName := "Product " + strconv.Itoa(i)
		itemDesc := "Test Description"
		itemQuality := "common"
		a.DB.Exec("INSERT INTO items(name, description, quality) VALUES($1, $2, $3)", itemName, itemDesc, itemQuality)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS items
(
    id SERIAL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
	quality TEXT NOT NULL,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/items", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestNonExistentItem(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/item/123", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Item not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Item not found'. Got '%s'", m["error"])
	}
}

func TestCreateItem(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"name":"Test Item", "description":"Test Description", "quality":"common"}`)
	req, _ := http.NewRequest("POST", "/item", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "Test Item" {
		t.Errorf("Expected item name to be 'Test Item'. Got '%v'", m["name"])
	}

	if m["description"] != "Test Description" {
		t.Errorf("Expected item description to be 'Test Description'. Got '%v'", m["description"])
	}

	if m["quality"] != "common" {
		t.Errorf("Expected item quality to be 'common'. Got '%v'", m["quality"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetItem(t *testing.T) {
	clearTable()
	addItems(1)

	req, _ := http.NewRequest("GET", "/item/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateItem(t *testing.T) {
	clearTable()
	addItems(1)

	req, _ := http.NewRequest("GET", "/item/1", nil)
	response := executeRequest(req)
	var originalItem map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalItem)

	var jsonStr = []byte(`{"name":"Updated Test Item", "description":"Updated Test Description", "quality": "rare"}`)
	req, _ = http.NewRequest("PUT", "/item/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalItem["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalItem["id"], m["id"])
	}

	if m["name"] == originalItem["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalItem["name"], "Updated Test Item", m["name"])
	}

	if m["description"] == originalItem["description"] {
		t.Errorf("Expected the description to change from '%v' to '%v'. Got '%v'", originalItem["description"], "Updated Test Description", m["description"])
	}

	if m["quality"] == originalItem["quality"] {
		t.Errorf("Expected the quality to change from '%v' to '%v'. Got '%v'", originalItem["quality"], "rare", m["quality"])
	}
}

func TestDeleteItem(t *testing.T) {
	clearTable()
	addItems(1)

	req, _ := http.NewRequest("GET", "/item/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/item/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/item/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
