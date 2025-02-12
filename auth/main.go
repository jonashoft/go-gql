package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"graphql-go/persistence"
	"net/url"
	"os"
	"time"

	"crypto/rand"
	"encoding/base64"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

var oauth2Config = &oauth2.Config{
	ClientID:     os.Getenv("AZURE_CLIENT_ID"),
	ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("AZURE_REDIRECT_URL"),
	Scopes:       []string{"https://graph.microsoft.com/.default"},
	Endpoint:     microsoft.AzureADEndpoint(os.Getenv("AZURE_TENANT_ID")),
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = 365 * 24 * 60 * 60

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, MaxAge: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	// Extract the redirect_url from query parameters
	redirectURL := r.URL.Query().Get("redirect_url")

	// Validate redirectURL here if necessary

	// Store redirect_url in a secure, http-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_redirect_url",
		Value:    url.QueryEscape(redirectURL), // Ensure the URL is safely encoded
		MaxAge:   300,                          // 5 minutes or adjust to your flow's timing requirements
		HttpOnly: true,                         // Make the cookie inaccessible to JavaScript
		Secure:   true,                         // Ensure cookie is sent over HTTPS only
		Path:     "/",                          // Scope of the cookie
		// Set other cookie attributes as necessary, such as SameSite
	})

	authCodeURL := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authCodeURL, http.StatusTemporaryRedirect)
}
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Retrieve the redirect_url from the cookie
	cookie, err := r.Cookie("oauth_redirect_url")
	if err != nil {
		// Handle missing or invalid cookie
		http.Error(w, "Session error", http.StatusBadRequest)
		return
	}
	// Decode the URL value from the cookie
	redirectURL, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		// Handle error
		http.Error(w, "Invalid session data", http.StatusBadRequest)
		return
	}
	// Read the oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, error2 := getUserInfo(token)
	if error2 != nil {
		http.Error(w, "Failed to get user info: "+error2.Error(), http.StatusInternalServerError)
		return
	}

	dbUser, error3 := upsertUser(user)

	if error3 != nil {
		http.Error(w, "Failed to upsert user: "+error3.Error(), http.StatusInternalServerError)
		return
	}
	// sign a jwt token
	jwtString, error4 := SignToken(dbUser)
	if error4 != nil {
		http.Error(w, "Failed to sign token: "+error4.Error(), http.StatusInternalServerError)
		return
	}
	// Construct the final redirect URL with the token
	finalRedirectURL := fmt.Sprintf("%s?token=%s", redirectURL, url.QueryEscape(jwtString))

	// Redirect the user to the frontend with the token
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}

func SignToken(user *persistence.User) (string, error) {
	// Define the token claims
	claims := jwt.MapClaims{
		"sub": user.ID,                                    // subject, you can use user's email or any unique identifier
		"exp": time.Now().Add(time.Hour * 24 * 60).Unix(), // token expiration time
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET"))) // Use your secret key here
	return tokenString, err
}

func upsertUser(user *User) (*persistence.User, error) {
	// Insert or update the user in the database
	gormDB := persistence.ConnectGORM()

	dbUser := &persistence.User{
		Email: user.Mail, // Use the email as the unique identifier for upsert
	}
	result := gormDB.Where(persistence.User{Email: user.Mail}).Attrs(persistence.User{
		ID:   uuid.New().String(), // Only set ID if creating a new record
		Name: user.DisplayName,
	}).FirstOrCreate(dbUser)

	return dbUser, result.Error
}

// User struct to hold the user information from Microsoft Graph
type User struct {
	DisplayName       string `json:"displayName"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"` // Often used as the user's email
}

func getUserInfo(token *oauth2.Token) (*User, error) {
	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, fmt.Errorf("request to Graph API failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Graph API returned non-200 status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding response body failed: %v", err)
	}

	return &user, nil
}
