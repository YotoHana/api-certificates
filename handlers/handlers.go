package handlers

import (
	config "api-certificates/configs"
	"api-certificates/structs"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

var conf = config.New()
var ctx = context.Background()
var secretKey = conf.SecretKey
var dbConn = conf.DbConn

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var data structs.Data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := pgx.Connect(ctx, dbConn)
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

		claims := structs.MyClaims{
			RegisteredClaims: jwt.RegisteredClaims{},
			Login: data.Login,
			Id: id,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		StrToken, err := token.SignedString(secretKey)
		if err != nil {
			log.Fatalf("Произошла ошибка %v\n", err)
		}

		resp := structs.Token{Token: StrToken}
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
		return secretKey, nil
	}

	claims := &structs.MyClaims{}
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
	
	conn, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		log.Fatalf("Unable to connect to database : %v\n", err)
	}
	defer conn.Close(ctx)

	query := "select age ->> 'title', amount ->> 'title' from filters_mapping"
	var age structs.Age
	var amount structs.Amount
	err = conn.QueryRow(ctx, query).Scan(&age.Title, &amount.Title)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}
	ageS := structs.Age{
		Title: age.Title,
	}
	amountS := structs.Amount{
		Title: amount.Title,
	}

	query = "select meta, filters_order from meta"
	var meta structs.Meta
	var filter structs.DataGrants
	err = conn.QueryRow(ctx, query).Scan(&meta, &filter.FiltersOrders)
	if err != nil {
		log.Fatalf("error queryrow: %v", err)
	}


	var projDir structs.ProjectDirection
	var legalForm structs.LegalForm
	var cutOfCrit structs.CuttingOffCriteria
	query = "select project_direction, legal_form, cutting_off_criteria from filters_mapping"
	err = conn.QueryRow(ctx, query).Scan(&projDir, &legalForm, &cutOfCrit)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}

	filterMapping := structs.FiltersMapping{
		Age: ageS,
		Amount: amountS,
		ProjectDirection: projDir,
		LegalForm: legalForm,
		CuttingOffCriteria: cutOfCrit,
	}

	dataGrants := structs.DataGrants{
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

	var grant structs.GrantItem
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func GrantsId(w http.ResponseWriter, r *http.Request)  {
	idString := r.PathValue("id")
	
	conn, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		log.Fatalf("Unable to connect to database : %v\n", err)
	}
	defer conn.Close(ctx)

	query := "select * from grants where id = $1"
	var grant structs.GrantItem
	err = conn.QueryRow(ctx, query, idString).Scan(
		&grant.ID,
			&grant.Title,
			&grant.SourceURL,
			&grant.FilterValues.ProjectDirection,
			&grant.FilterValues.Amount,
			&grant.FilterValues.LegalForm,
			&grant.FilterValues.Age,
			&grant.FilterValues.CuttingOffCriteria,
	)
	if err != nil {
		log.Fatalf("Error with queryRow: %v", err)
	}

	query = "select age ->> 'title', amount ->> 'title' from filters_mapping"
	var age structs.Age
	var amount structs.Amount
	err = conn.QueryRow(ctx, query).Scan(&age.Title, &amount.Title)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}
	ageS := structs.Age{
		Title: age.Title,
	}
	amountS := structs.Amount{
		Title: amount.Title,
	}

	query = "select filters_order from meta"
	var filter structs.DataGrants
	err = conn.QueryRow(ctx, query).Scan(&filter.FiltersOrders)
	if err != nil {
		log.Fatalf("error queryrow: %v", err)
	}

	var projDir structs.ProjectDirection
	var legalForm structs.LegalForm
	var cutOfCrit structs.CuttingOffCriteria
	query = "select project_direction, legal_form, cutting_off_criteria from filters_mapping"
	err = conn.QueryRow(ctx, query).Scan(&projDir, &legalForm, &cutOfCrit)
	if err != nil {
		log.Fatalf("Error queryrow: %v", err)
	}

	filterMapping := structs.FiltersMapping{
		Age: ageS,
		Amount: amountS,
		ProjectDirection: projDir,
		LegalForm: legalForm,
		CuttingOffCriteria: cutOfCrit,
	}

	dataGrants := structs.DataGrantItem{
		Grants: grant,
		FiltersMapping: filterMapping,
		FiltersOrders: filter.FiltersOrders,
	}

	jsonResult, err := json.Marshal(dataGrants)
	if err != nil {
		log.Fatalf("Error with marshall: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func GrantsFilters(w http.ResponseWriter, r *http.Request)  {
	idString := r.PathValue("id")
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		log.Fatalf("Error with strconv %v", err)
	}
	
	var respDataFilters structs.DataFilters
	err = json.NewDecoder(r.Body).Decode(&respDataFilters)
	if err != nil {
		log.Fatalf("Error with decoding JSON: %v", err)
	}
	
	query := "UPDATE grants SET project_directions = $1, amount = $2, legal_forms = $3, age = $4, cutting_off_criterea = $5 where id = $6"
	conn, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		log.Fatalf("Error with database: %v", err)
	}
	updateRow, err := conn.Exec(ctx, query, &respDataFilters.Data.ProjectDirection, &respDataFilters.Data.Amount, &respDataFilters.Data.LegalForm, &respDataFilters.Data.Age, &respDataFilters.Data.CuttingOffCriteria, idInt)
	if err != nil {
		log.Fatalf("Error with queryRow : %v", err)
	}
	if updateRow.RowsAffected() != 1 {
		log.Fatalf("No row found to updated")
	}

	w.WriteHeader(http.StatusNoContent)

}