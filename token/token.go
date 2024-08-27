package token

import (
	config "api-certificates/configs"
	"api-certificates/structs"
	"encoding/json"
	"fmt"
	"log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

var conf = config.New()
var secretKey = []byte(conf.SecretKey)

func New(JWTstruct *structs.MyClaims) (jsonData []byte, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTstruct)
	t, err := token.SignedString(secretKey)
	if err != nil {
		return nil, fmt.Errorf("Error with signed JWT: %v", err)
	}
	resp := structs.Token{Token: t}
	jsonData, err = json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("Error with marshalling JSON: %v", err)
	}
	return jsonData, nil
}

func Validate(tokenString string) (bool, error) {
	method := func (t *jwt.Token) (interface{}, error)  {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil 
	}

	claims := &structs.MyClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, method)
	if err != nil {
		return false, fmt.Errorf("Error with parse JWT: %v", err)
	}
	if !parsedToken.Valid {
		return false, fmt.Errorf("JWT is not valid")
	}
	return true , nil
}