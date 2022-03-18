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
			   http://localhost:8010/graphql/item?query={item(id:1){name,description,quality}}
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

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		/* Create new item
		http://localhost:8010/graphql/item?query=mutation{create(name:"Crowbar",description:"Deal +75% (+75% per stack) damage to enemies above 90% health.",quality:"common"){id,name,description,quality}}
		*/
		"create": &graphql.Field{
			Type:        itemType,
			Description: "Create new item",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"quality": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				db := p.Context.Value("db").(*sql.DB)
				i := item{
					Name:        p.Args["name"].(string),
					Description: p.Args["description"].(string),
					Quality:     p.Args["quality"].(string),
				}
				err := i.createItem(db)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
		},
		/* Update item by id
		http://localhost:8010/graphql/item?query=mutation{update(id:1,quality:"white"){id,name,description,quality}}
		*/
		"update": &graphql.Field{
			Type:        itemType,
			Description: "Update item by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"quality": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				db := p.Context.Value("db").(*sql.DB)
				id, _ := p.Args["id"].(int)
				i := item{
					ID: id,
				}
				err := i.getItem(db)
				if err != nil {
					return nil, err
				}

				name, nameOk := p.Args["name"].(string)
				if nameOk {
					i.Name = name
				}
				description, descriptionOk := p.Args["description"].(string)
				if descriptionOk {
					i.Description = description
				}
				quality, qualityOk := p.Args["quality"].(string)
				if qualityOk {
					i.Quality = quality
				}

				err = i.updateItem(db)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
		},
		/* Delete item by id
		http://localhost:8010/graphql/item?query=mutation{delete(id:1){id,name,description,quality}}
		*/
		"delete": &graphql.Field{
			Type:        itemType,
			Description: "Delete item by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				db := p.Context.Value("db").(*sql.DB)
				id, _ := p.Args["id"].(int)
				i := item{
					ID: id,
				}

				err := i.deleteItem(db)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
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
