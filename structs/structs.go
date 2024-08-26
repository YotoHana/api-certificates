package structs

import "github.com/golang-jwt/jwt/v5"

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

type DataFilters struct {
	Data struct {
		ProjectDirection []int `json:"project_direction"`
		Amount int `json:"amount"`
		LegalForm []int `json:"legal_form"`
		Age int `json:"age"`
		CuttingOffCriteria []int `json:"cutting_off_criteria"`
	} `json:"data"`
}

type DataGrants struct {
	Grants []GrantItem `json:"grants"`
	FiltersMapping FiltersMapping `json:"filters_mapping"`
	FiltersOrders []string `json:"filter_order"`
	Meta Meta `json:"meta,omitempty"`
}

type DataGrantItem struct {
	Grants GrantItem `json:"grant"`
	FiltersMapping FiltersMapping `json:"filters_mapping"`
	FiltersOrders []string `json:"filter_order"`
}

type GrantItem struct {
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
	Mapping MappingThree `json:"mapping"`
}

type LegalForm struct {
	Title string `json:"title"`
	Mapping MappingTwo `json:"mapping"`
}

type CuttingOffCriteria struct{
	Title string `json:"title"`
	Mapping MappingThree `json:"mapping"`
}

type Amount struct {
	Title string `json:"title"`
	Mapping MappingEmpty `json:"mapping"`
}

type MappingEmpty struct {
	Empty string `json:"empty,omitempty"`
}

type MappingThree struct {
	Zero struct{
		Title string `json:"title"`
	} `json:"0"`
	One struct{
		Title string `json:"title"`
	} `json:"1"`
	Two struct{
		Title string `json:"title"`
	} `json:"2"`
	Three struct{
		Title string `json:"title,omitempty"`
	} `json:"3,omitempty"`
}

type MappingTwo struct {
	Zero struct{
		Title string `json:"title"`
	} `json:"0"`
	One struct{
		Title string `json:"title"`
	} `json:"1"`
	Two struct{
		Title string `json:"title"`
	} `json:"2"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	TotalPages int `json:"total_pages"`
}