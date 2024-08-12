package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type Data struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Token struct {
	Token string `json:"token"`
}

type MyClaims struct {
	jwt.RegisteredClaims
	Id int `json:"id"`
	Login string `json:"login"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var data Data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := pgx.Connect(context.Background(),"postgres://adminfront:123@localhost:5432/adminfront")
	if err != nil {
		fmt.Fprintf(os.Stderr,"Unable to connect to database : %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var password_match bool
	err = conn.QueryRow(context.Background(),"SELECT (password = crypt($1, password)) as password_match FROM users WHERE login = $2", data.Password, data.Login).Scan(&password_match)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Password_match: %v", password_match)

	if password_match {
		var id int
		err = conn.QueryRow(context.Background(),"SELECT id FROM users WHERE login = $1", data.Login).Scan(&id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			os.Exit(1)
		}

		claims := MyClaims{
			RegisteredClaims: jwt.RegisteredClaims{},
			Login: data.Login,
			Id: id,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		var mySigningKey = []byte("secret-key")

		StrToken, err := token.SignedString(mySigningKey)
		if err != nil {
			log.Fatalf("Произошла ошибка %v\n", err)
		}

		resp := Token{Token: StrToken}
		log.Println("Login data is correct.")
		jsonData, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Login data is not correct. Error!")
	}
}

func CheckHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == StrToken{
		w.WriteHeader(http.StatusNoContent)
	}
	log.Println("Authorization token is sending...")
}

// func GrantsHandler(w http.ResponseWriter, r *http.Request)  {
// 	header := r.FormValue("page")
// 	log.Printf("%s Page", header)
// }

func main() {
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/check", CheckHandler)
	// http.HandleFunc("/grants", GrantsHandler)
	log.Println("Server started at localhost:8080")
	http.ListenAndServe(":8080", nil)
}