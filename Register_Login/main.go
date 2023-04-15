package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"test/helpers"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

const (
	dbHost     = "localhost"
	dbPort     = "5432"
	dbUser     = "postgres"
	dbPassword = "321654"
	dbName     = "postgres"
)

type User struct {
	ID int `json:"id"`
	// Username string `json:"username"`
	// Password        string `json:"-"`
	PasswordConfirm string `json:"--"`
	Email           string `json:"email"`
	Claims          *Claims
}

// type Credentials struct {
// 	Claims   *Claims
// 	// Password string `json:"password"`
// }

type Claims struct {
	LoginRequest *LoginRequest
	jwt.StandardClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var tokenString string

func main() {
	// Connect to database
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	var err error
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/signup", SignupHandler).Methods("POST")
	router.HandleFunc("/login", LoginHandler).Methods("POST")

	// Start server
	log.Println("Server started on port 3000")
	err = http.ListenAndServe(":3000", router)
	if err != nil {
		log.Fatal(err)
	}
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	http.ServeFile(w, r, "register.html")
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if username or password is empty
	if user.Claims.LoginRequest.Username == "" || user.Claims.LoginRequest.Password == "" {
		http.Error(w, "Username and password cannot be empty.", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	usertemp, err := helpers.GetUserByUsername(user.Claims.LoginRequest.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if usertemp != nil {
		http.Error(w, "User already exists.", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Claims.LoginRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert user into database
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Claims.LoginRequest.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create and sign JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		user.Claims.LoginRequest, //------------------------------------------------ BURASINDA HATA OLABİLİR
		jwt.StandardClaims{
			ExpiresAt: 15000, // Token expires in 15 seconds
			Issuer:    "myapp",
		},
	})
	tokenString, err = token.SignedString([]byte("mysecret"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return token

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	http.ServeFile(w, r, "login.html")
	decoder := json.NewDecoder(r.Body)
	var loginRequest LoginRequest
	err := decoder.Decode(&loginRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if username or password is empty
	if helpers.IsEmpty(loginRequest.Username) || helpers.IsEmpty(loginRequest.Password) {
		http.Error(w, "Username and password cannot be empty.", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := helpers.GetUserByUsername(loginRequest.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid username or password.", http.StatusUnauthorized)
		return
	}

	// Compare hashed password with plain password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		http.Error(w, "Invalid username or password.", http.StatusUnauthorized)
		return
	}

	// Create and sign JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": loginRequest.Username,
	})
	tokenString, err = token.SignedString([]byte("secret"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return token to user
	loginResponse := LoginResponse{
		Token: tokenString,
	}
	jsonResponse, err := json.Marshal(loginResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
