package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/graphql-go/graphql"
	_ "github.com/lib/pq"
)

var itemType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Item",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"quality": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			/* Get (read) single item by id
			   http://localhost:8010/item?query={item(id:1){name,description,quality}}
			*/
			"item": &graphql.Field{
				Type:        itemType,
				Description: "Get item by id",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if ok {
						// find item
						i := item{ID: id}
						if err := i.getItem(p.Context.Value("db").(*sql.DB)); err != nil {
							return nil, err
						}
						return i, nil
					}
					return nil, nil
				},
			},
			/* Get (read) item list
			   http://localhost:8010/graphql/item?query={list{id,name,description,quality}}
			*/
			"list": &graphql.Field{
				Type:        graphql.NewList(itemType),
				Description: "Get item list",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					items, err := getItems(p.Context.Value("db").(*sql.DB))
					if err != nil {
						return nil, err
					}
					return items, nil
				},
			},
		},
	},
)

// implement mutation types here (create/update/delete)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
		//Mutation: mutationType,
	},
)

func executeQuery(query string, schema graphql.Schema, db *sql.DB) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "db", db),
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}
