package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"what-to-eat/be/auth"
	"what-to-eat/be/controllers"
	"what-to-eat/be/directive"
	"what-to-eat/be/firebase"
	"what-to-eat/be/graph"
	"what-to-eat/be/shared"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/handlers"
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

	router := mux.NewRouter().StrictSlash(true)

	dishRouter := router.PathPrefix("/dish").Subrouter()
	dishRouter.HandleFunc("/", controllers.NewDishController().Find).Methods("GET")
	dishRouter.HandleFunc("/random", controllers.NewDishController().FindRandom).Methods("GET")
	dishRouter.HandleFunc("/{slug}", controllers.NewDishController().FindOne).Methods("GET")

	ingredientRouter := router.PathPrefix("/ingredient").Subrouter()
	ingredientRouter.HandleFunc("/", controllers.NewIngredientController().Find).Methods("GET")
	ingredientRouter.HandleFunc("/{slug}", controllers.NewIngredientController().FindOne).Methods("GET")

	authRouter := router.Methods(http.MethodPost, http.MethodGet).Subrouter()
	authRouter.HandleFunc("/login", auth.NewAuthController().Login)
	authRouter.HandleFunc("/retrieve-token", auth.NewAuthController().RetrieveToken)

	c := graph.Config{Resolvers: &graph.Resolver{}}
	c.Directives.HasRole = directive.Role

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))

	graphRouter := router.Methods(http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete).Subrouter()
	graphRouter.Use(auth.Middleware())

	authRouter.Handle("/", playground.Handler("GraphQL playground", "/query"))
	graphRouter.Handle("/query", srv)

	headersOk := handlers.AllowedHeaders([]string{"Access-Control-Allow-Origin", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(headersOk, originsOk, methodsOk)(router)))
}
