package main

import (
	"database/sql"
)

type item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quality     string `json:"quality"`
}

func (i *item) getItem(db *sql.DB) error {
	return db.QueryRow("SELECT name, description, quality FROM items WHERE id=$1", i.ID).Scan(&i.Name, &i.Description, &i.Quality)
}

func (i *item) createItem(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO items(name, description, quality) VALUES($1, $2, $3) RETURNING id", i.Name, i.Description, i.Quality).Scan(&i.ID)

	if err != nil {
		return err
	}

	return nil
}

func (i *item) updateItem(db *sql.DB) error {
	_, err := db.Exec("UPDATE items SET name=$1, description=$2, quality=$3 WHERE id=$4", i.Name, i.Description, i.Quality, i.ID)

	return err
}

func (i *item) deleteItem(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM items WHERE id=$1", i.ID)

	return err
}

func getItems(db *sql.DB) ([]item, error) {
	rows, err := db.Query("SELECT id, name, description, quality FROM items LIMIT 1000")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []item{}
	for rows.Next() {
		var i item
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.Quality); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, nil
}
