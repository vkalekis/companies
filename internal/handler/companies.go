package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vkalekis/companies/internal/cdc"
	"github.com/vkalekis/companies/pkg/model"
)

func (h *Handler) setError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	data, err := json.Marshal(model.Error{
		Err: err.Error(),
	})
	if err != nil {
		h.log.Printf("Error in JSON marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("Error writing data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetCompanies(w http.ResponseWriter, r *http.Request) {

	rows, err := h.pool.Query(context.Background(), h.queries[GetCompaniesKey])
	if err != nil {
		h.log.Printf("GetCompanies: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var companies []model.Company
	var company model.Company
	var description sql.Null[string]

	for rows.Next() {
		err := rows.Scan(
			&company.Id,
			&company.Name,
			&description,
			&company.Employees,
			&company.Registered,
			&company.CompanyType,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			h.log.Printf("GetCompanies: Scan error: %v", err)
			continue
		}

		if description.Valid {
			company.Description = &description.V
		}
		companies = append(companies, company)
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(companies)
	if err != nil {
		h.log.Printf("GetCompanies: Error in JSON marshal: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("GetCompanies: Error writing data: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetCompany(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if idStr == "" || err != nil {
		h.log.Printf("GetCompany: Invalid id")
		h.setError(w, ErrInvalidArg{id: id}, http.StatusBadRequest)
		return
	}

	company, err := h.getCompany(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			h.setError(w, ErrNotFound{id: id}, http.StatusNotFound)
			return
		}
		h.log.Printf("GetCompany: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(company)
	if err != nil {
		h.log.Printf("GetCompany: Error in JSON marshal: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("GetCompany: Error writing data: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Printf("CreateCompany: unable to read body: %v", err)
		h.setError(w, ErrInvalidArg{}, http.StatusBadRequest)
		return
	}

	var req model.CompanyCreateReq
	if err := json.Unmarshal(body, &req); err != nil {
		h.log.Printf("CreateCompany: invalid json: %v", err)
		h.setError(w, ErrInvalidArg{body: body}, http.StatusBadRequest)
		return
	}

	if !h.areFieldsValid(req) {
		h.log.Printf("CreateCompany: invalid fields")
		h.setError(w, ErrInvalidArg{}, http.StatusBadRequest)
		return
	}

	var company model.Company
	var description sql.Null[string]

	err = h.pool.QueryRow(context.Background(),
		h.queries[CreateCompanyKey],
		uuid.New(),
		req.Name,
		req.Description,
		req.Employees,
		req.Registered,
		req.CompanyType,
	).Scan(
		&company.Id,
		&company.Name,
		&description,
		&company.Employees,
		&company.Registered,
		&company.CompanyType,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if description.Valid {
		company.Description = &description.V
	}

	if err != nil {
		h.log.Printf("CreateCompany: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	h.cdcChecker.Register(cdc.Operation{
		Before: nil,
		After:  &company,
		Op:     cdc.Op_Create,
	})

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(company)
	if err != nil {
		h.log.Printf("CreateCompany: Error in JSON marshal: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("CreateCompany: Error writing data: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if idStr == "" || err != nil {
		h.log.Printf("UpdateCompany: Invalid id")
		h.setError(w, ErrInvalidArg{id: id}, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Printf("UpdateCompany: unable to read body: %v", err)
		h.setError(w, ErrInvalidArg{}, http.StatusBadRequest)
		return
	}

	var req model.CompanyUpdateReq
	if err := json.Unmarshal(body, &req); err != nil {
		h.log.Printf("UpdateCompany: invalid json: %v", err)
		h.setError(w, ErrInvalidArg{body: body}, http.StatusBadRequest)
		return
	}

	if !h.areFieldsValid(req) {
		h.log.Printf("UpdateCompany: invalid fields")
		h.setError(w, ErrInvalidArg{}, http.StatusBadRequest)
		return
	}

	before, err := h.getCompany(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			h.setError(w, ErrNotFound{id: id}, http.StatusInternalServerError)
			return
		}
		h.log.Printf("UpdateCompany: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	var company model.Company
	var description sql.Null[string]

	err = h.pool.QueryRow(context.Background(),
		h.queries[UpdateCompanyKey],
		id,
		req.Name,
		req.Description,
		req.Employees,
		req.Registered,
		req.CompanyType,
	).Scan(
		&company.Id,
		&company.Name,
		&description,
		&company.Employees,
		&company.Registered,
		&company.CompanyType,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if description.Valid {
		company.Description = &description.V
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			h.setError(w, ErrNotFound{id: id}, http.StatusInternalServerError)
			return
		}
		h.log.Printf("UpdateCompany: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	h.cdcChecker.Register(cdc.Operation{
		Before: &before,
		After:  &company,
		Op:     cdc.Op_Update,
	})

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(company)
	if err != nil {
		h.log.Printf("CreateCompany: Error in JSON marshal: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("CreateCompany: Error writing data: %v", err)
		h.setError(w, FFormatResponseError, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteCompany(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if idStr == "" || err != nil {
		h.log.Printf("DeleteCompany: Invalid id")
		h.setError(w, ErrInvalidArg{id: id}, http.StatusBadRequest)
		return
	}

	befCompany, err := h.getCompany(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			h.setError(w, ErrNotFound{id: id}, http.StatusInternalServerError)
			return
		}
		h.log.Printf("DeleteCompany: query error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	_, err = h.pool.Exec(context.Background(),
		h.queries[DeleteCompanyKey],
		id,
	)
	if err != nil {
		h.log.Printf("DeleteCompany: exec error: %v", err)
		h.setError(w, QQueryErr, http.StatusInternalServerError)
		return
	}

	h.cdcChecker.Register(cdc.Operation{
		Before: &befCompany,
		After:  nil,
		Op:     cdc.Op_Delete,
	})
}

func (h *Handler) getCompany(id uuid.UUID) (model.Company, error) {
	var company model.Company
	var description sql.Null[string]

	err := h.pool.QueryRow(context.Background(),
		h.queries[GetCompanyKey], id,
	).Scan(
		&company.Id,
		&company.Name,
		&description,
		&company.Employees,
		&company.Registered,
		&company.CompanyType,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return model.Company{}, err
	}

	if description.Valid {
		company.Description = &description.V
	}

	return company, nil
}

// areFieldsValid checks if the fields of the given company are valid
//   - sane number of employees
//   - company type in one of the accepted categories
func (h *Handler) areFieldsValid(company interface{}) bool {

	var employees int
	var companyType model.CompanyType
	switch v := company.(type) {
	case model.CompanyCreateReq:
		employees = v.Employees
		companyType = v.CompanyType
	case model.CompanyUpdateReq:
		// Insert some dummy values to pass the validation checks
		if v.Employees == nil {
			employees = 1
		} else {
			employees = *v.Employees
		}

		if v.CompanyType == nil {
			companyType = model.CompanyType_Corporation
		} else {
			companyType = *v.CompanyType
		}

	default:
		return false
	}

	if employees < 0 {
		return false
	}

	switch companyType {
	case model.CompanyType_Corporation, model.CompanyType_Cooperative, model.CompanyType_NonProfit, model.CompanyType_SoleProprietorship:
	default:
		return false
	}

	return true
}
