package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vkalekis/companies/internal/auth"
	"github.com/vkalekis/companies/pkg/model"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	h.log.Printf("Hit Login")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Printf("Login: unable to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var login model.Login
	if err := json.Unmarshal(body, &login); err != nil {
		h.log.Printf("Login: invalid json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if login.Username == "" || login.Password == "" {
		h.log.Printf("Login: expected username and password to be non empty: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var dbId uuid.UUID

	err = h.pool.QueryRow(context.Background(),
		h.queries[LoginKey],
		login.Username,
		login.Password,
	).Scan(
		&dbId,
	)

	h.log.Printf("Login: User %s/%s logged in", login.Username, dbId)

	if err != nil {
		if err == pgx.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.log.Printf("Login: query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwt, err := auth.CreateJWT(login.Username)
	if err != nil {
		h.log.Printf("Login: create jwt error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(
		model.JWT{
			JWT: jwt,
		},
	)
	if err != nil {
		h.log.Printf("Login: Error in JSON marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.log.Printf("Login: Error writing data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
