package model

import (
	"time"

	"github.com/google/uuid"
)

type CompanyType string

const (
	CompanyType_Corporation        = "Corporations"
	CompanyType_NonProfit          = "NonProfit"
	CompanyType_Cooperative        = "Cooperative"
	CompanyType_SoleProprietorship = "Sole Proprietorship"
)

type Company struct {
	Id          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Employees   int         `json:"employees"`
	Registered  bool        `json:"registered"`
	CompanyType CompanyType `json:"company_type"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type CompanyCreateReq struct {
	Name        string      `json:"name"`
	Description *string     `json:"description"`
	Employees   int         `json:"employees"`
	Registered  bool        `json:"registered"`
	CompanyType CompanyType `json:"company_type"`
}

type CompanyUpdateReq struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	Employees   *int         `json:"employees"`
	Registered  *bool        `json:"registered"`
	CompanyType *CompanyType `json:"company_type"`
}
