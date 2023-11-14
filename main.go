package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"what-to-eat/be/auth"
	"what-to-eat/be/directive"
	"what-to-eat/be/firebase"
	"what-to-eat/be/graph"
	"what-to-eat/be/shared"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	shared.InitializeMongoDB()
	firebase.InitFirebase()

	router := mux.NewRouter()
	router.Use(auth.Middleware())
	c := graph.Config{Resolvers: &graph.Resolver{}}
	c.Directives.Auth = directive.Auth
	c.Directives.HasRole = directive.Role

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
