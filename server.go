package main

import (
	"graphql-go/db"
	"graphql-go/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

// Middleware to enforce nonce check
func nonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for nonce in query parameters
		nonce := r.URL.Query().Get("nonce")
		// For demonstration, let's assume the valid nonce is "12345"
		validNonce := os.Getenv("VALID_NONCE")

		// Validate the nonce
		if nonce != validNonce {
			// Respond with an error if the nonce is invalid
			http.Error(w, "Invalid or missing nonce", http.StatusUnauthorized)
			return
		}

		// If the nonce is valid, proceed with the request
		next.ServeHTTP(w, r)
	})
}
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create your GraphQL server handler
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB: db.Connect(),
	}}))

	// Wrap the /query handler with nonceMiddleware to protect your GraphQL API
	http.Handle("/query", nonceMiddleware(srv))

	// Optionally, protect the GraphQL playground with the nonceMiddleware as well
	// This means accessing the playground will also require a valid nonce
	playgroundHandler := playground.Handler("GraphQL playground", "/query")
	http.Handle("/", nonceMiddleware(playgroundHandler))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
