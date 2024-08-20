package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

type DataGrants struct {
	Grants []Grant `json:"grants"`
}

type Grant struct {
	ID           int          `json:"id" db:"id"`
	Title        string       `json:"title" db:"title"`
	SourceURL    string       `json:"source_url" db:"source_url"`
	FilterValues FilterValues `json:"filter_values"`
}

type FilterValues struct {
	CuttingOffCriteria []int `json:"cutting_off_criteria" db:"cutting_off_criterea"`
	ProjectDirection   []int `json:"project_direction" db:"project_directions"`
	Amount             int   `json:"amount" db:"amount"`
	LegalForm          []int `json:"legal_form" db:"legal_forms"`
	Age                int   `json:"age" db:"age"`
}

var mySigningKey = []byte("secret-key")
var ctx = context.Background()

func Login(w http.ResponseWriter, r *http.Request) {
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

	conn, err := pgx.Connect(ctx,"postgres://adminfront:123@localhost:5432/adminfront")
	if err != nil {
		log.Fatalf("Unable to connect to database : %v\n", err)
	}
	defer conn.Close(ctx)

	var password_match bool
	err = conn.QueryRow(ctx,"SELECT (password = crypt($1, password)) as password_match FROM users WHERE login = $2", data.Password, data.Login).Scan(&password_match)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}

	log.Printf("Password_match: %v", password_match)

	if password_match {
		var id int
		err = conn.QueryRow(ctx,"SELECT id FROM users WHERE login = $1", data.Login).Scan(&id)
		if err != nil {
			log.Fatalf("QueryRow failed: %v\n", err)
		}

		claims := MyClaims{
			RegisteredClaims: jwt.RegisteredClaims{},
			Login: data.Login,
			Id: id,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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
	}
}

func Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Неожиданный метод подписи: %v", t.Header["alg"])
		}
		return mySigningKey, nil
	}

	claims := &MyClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		log.Fatalf("Ошибка разбора check: %v", err)
	}

	if !parsedToken.Valid {
		log.Fatalf("Недействительный токен")
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func Grants (w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	conn, err := pgx.Connect(ctx,"postgres://adminfront:123@localhost:5432/adminfront")
	if err != nil {
		log.Fatalf("Unable to connect to database : %v\n", err)
	}
	defer conn.Close(ctx)

	query := "SELECT id, title, source_url, project_directions, amount, legal_forms, age, cutting_off_criterea FROM grants WHERE id = $1"
    var grant Grant

    // Выполнение запроса с привязкой параметра id
    err = conn.QueryRow(context.Background(), query, 1).Scan(
        &grant.ID,
        &grant.Title,
        &grant.SourceURL,
        // Здесь нужно будет преобразовать или десериализовать значения, если они JSON
        &grant.FilterValues.ProjectDirection, // предполагаем, что это массив
        &grant.FilterValues.Amount,
        &grant.FilterValues.LegalForm, // если это массив
        &grant.FilterValues.Age,
        &grant.FilterValues.CuttingOffCriteria, // если это массив
    )
    
    if err != nil {
        log.Fatal(err)
    }

    // Создание объекта DataGrants
    dataGrants := DataGrants{
        Grants: []Grant{grant},
    }

    // Преобразование результата в JSON
    jsonResult, err := json.Marshal(dataGrants)
    if err != nil {
        log.Fatal(err)
    }

    // Вывод JSON
    fmt.Println(string(jsonResult))

	
}