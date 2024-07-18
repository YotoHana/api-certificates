package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Data struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type Response struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type Token struct {
	Token string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request){
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

	if data.Login == "admin" && data.Password == "correct_password" {
		resp := Token{Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHAiOiIyMDIzLTEyLTE4VDEyOjI5OjE5LjEwNjg0MTQzOVoiLCJVc2VyTG9naW4iOiJhZG1pbiJ9.0Dvg7vFTrdSX2F4751ae6Id9weC5ATvF1sQPuvejiFE"}
		jsonData, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		resp := Response{Code: 401, Message: "Unauthorized"}
		jsonData, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func CheckHandler(w http.ResponseWriter, r *http.Request)  {
	token := r.Header.Get("Authorization")
	if token == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHAiOiIyMDIzLTEyLTE4VDEyOjI5OjE5LjEwNjg0MTQzOVoiLCJVc2VyTG9naW4iOiJhZG1pbiJ9.0Dvg7vFTrdSX2F4751ae6Id9weC5ATvF1sQPuvejiFE" {
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
 http.HandleFunc("/admin/api/v1/auth/login", LoginHandler)
 http.HandleFunc("/admin/api/v1/auth/check", CheckHandler)
 log.Println("Server started at localhost:80")
 http.ListenAndServe(":80", nil)
}