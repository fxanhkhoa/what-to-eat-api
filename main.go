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

type contextKey struct {
	name string
}

const defaultPort = "8080"

var userCtxKey = &contextKey{"user"}

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

	authRouter := router.Methods(http.MethodPost).Subrouter()
	authRouter.HandleFunc("/login", auth.NewAuthController().Login)
	authRouter.HandleFunc("/retrieve-token", auth.NewAuthController().RetrieveToken)

	c := graph.Config{Resolvers: &graph.Resolver{}}
	c.Directives.HasRole = directive.Role

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))

	graphRouter := router.Methods(http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete).Subrouter()
	graphRouter.Use(auth.Middleware())
	graphRouter.Handle("/", playground.Handler("GraphQL playground", "/query"))
	graphRouter.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
