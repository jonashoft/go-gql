package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"graphql-go/auth"
	"graphql-go/graph"
	"graphql-go/persistence"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "localhost:"+defaultPort {
			next.ServeHTTP(w, r)
			return
		}

		// Check for JWT in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// The header should be in the format `Bearer <token>`
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate the JWT
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Make sure the token method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// Return the secret key (this should be more secure in production)
			// Replace "your_secret_key" with your actual secret key
			return []byte("your_secret_key"), nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// You can access the token claims here if needed
			fmt.Println(claims)
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// If the JWT is valid, proceed with the request
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Nonce")

		// If the request is an OPTIONS request, return immediately with a 200 response
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue processing the request
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
		DB: persistence.ConnectGORM(),
	}}))

	// Wrap the /query handler with nonceMiddleware to protect your GraphQL API
	http.Handle("/query", corsMiddleware(authMiddleware(srv)))
	http.HandleFunc("/login", auth.HandleLogin)
	http.HandleFunc("/auth/callback", auth.HandleCallback)

	// Optionally, protect the GraphQL playground with the nonceMiddleware as well
	// This means accessing the playground will also require a valid nonce
	playgroundHandler := playground.Handler("GraphQL playground", "/query")
	http.Handle("/", playgroundHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
