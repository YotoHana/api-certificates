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
	FiltersMapping FiltersMapping `json:"filters_mapping"`
	FiltersOrders []string `json:"filter_order"`
	Meta Meta `json:"meta"`
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

type FiltersMapping struct {
	Age Age `json:"age"`
	ProjectDirection ProjectDirection `json:"project_direction"`
	LegalForm LegalForm `json:"legal_form"`
	CuttingOffCriteria CuttingOffCriteria `json:"cutting_off_criteria"`
	Amount Amount `json:"amount"`
}

type Age struct {
	Title string `json:"title"`
	Mapping MappingEmpty `json:"mapping,omitempty"`
}

type ProjectDirection struct {
	Title string `json:"title"`
	Mapping Mapping `json:"mapping"`
}

type LegalForm struct {
	Title string `json:"title"`
	Mapping Mapping `json:"mapping"`
}

type CuttingOffCriteria struct{
	Title string `json:"title"`
	Mapping Mapping `json:"mapping"`
}

type Amount struct {
	Title string `json:"title"`
	Mapping MappingEmpty `json:"mapping"`
}

type MappingEmpty struct {
	Empty string `json:"empty,omitempty"`
}

type Mapping struct {
	Zero Zero `json:"0"`
	Two Two `json:"1"`
	Three Three `json:"2"`
}

type Zero struct {
	Title string `json:"title"`
}

type Two struct {
	Title string `json:"title"`
}

type Three struct {
	Title string `json:"title"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	TotalPages int `json:"total_pages"`
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

	query := "select age ->> 'title', amount ->> 'title' from filters_mapping"
	var age Age
	var amount Amount
	err = conn.QueryRow(ctx, query).Scan(&age.Title, &amount.Title)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}
	ageS := Age{
		Title: age.Title,
	}
	amountS := Amount{
		Title: amount.Title,
	}

	query = "select meta, filters_order from meta"
	var meta Meta
	var filter DataGrants
	err = conn.QueryRow(ctx, query).Scan(&meta, &filter.FiltersOrders)
	if err != nil {
		log.Fatalf("error queryrow: %v", err)
	}


	var projDir ProjectDirection
	var legalForm LegalForm
	var cutOfCrit CuttingOffCriteria
	query = "select project_direction, legal_form, cutting_off_criteria from filters_mapping"
	err = conn.QueryRow(ctx, query).Scan(&projDir, &legalForm, &cutOfCrit)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}

	filterMapping := FiltersMapping{
		Age: ageS,
		Amount: amountS,
		ProjectDirection: projDir,
		LegalForm: legalForm,
		CuttingOffCriteria: cutOfCrit,
	}

	dataGrants := DataGrants{
		Grants: nil,
		FiltersMapping: filterMapping,
		Meta: meta,
		FiltersOrders: filter.FiltersOrders,
	}

	query = "SELECT * FROM grants"

    rows, err := conn.Query(ctx, query)
	if err != nil {
		log.Fatalf("Error query: %v",err)
	}
	defer rows.Close()

	var grant Grant
	for rows.Next() {
		rows.Scan(
			&grant.ID,
			&grant.Title,
			&grant.SourceURL,
			&grant.FilterValues.ProjectDirection,
			&grant.FilterValues.Amount,
			&grant.FilterValues.LegalForm,
			&grant.FilterValues.Age,
			&grant.FilterValues.CuttingOffCriteria,
		)
			dataGrants.Grants = append(dataGrants.Grants, grant)
	}

	if rows.Err() != nil {
        log.Fatalf("error rows query: %v",err)
    }

    jsonResult, err := json.Marshal(dataGrants)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(jsonResult))

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)

}