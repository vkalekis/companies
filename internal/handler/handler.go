package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vkalekis/companies/internal/auth"
	"github.com/vkalekis/companies/internal/cdc"
	"github.com/vkalekis/companies/internal/config"
)

type Handler struct {
	config     *config.Config
	log        *log.Logger
	pool       *pgxpool.Pool
	queries    map[string]string
	httpServer *http.Server
	cdcChecker *cdc.Checker
}

func New(config *config.Config, pool *pgxpool.Pool, cdcOperator cdc.Operator) *Handler {
	log := log.New(os.Stdout, "[handler] ", log.Ldate|log.Ltime|log.Lshortfile)
	return &Handler{
		config: config,
		log:    log,
		pool:   pool,
		queries: map[string]string{
			LoginKey:         LoginKeyQuery,
			GetCompaniesKey:  GetCompaniesQuery,
			GetCompanyKey:    GetCompanyQuery,
			CreateCompanyKey: CreateCompanyQuery,
			UpdateCompanyKey: UpdateCompanyQuery,
			DeleteCompanyKey: DeleteCompanyQuery,
		},
		cdcChecker: cdc.NewChecker(cdcOperator),
	}
}

func (h *Handler) Start() {

	r := http.NewServeMux()

	r.Handle("POST /login", h.loggingMiddleware(http.HandlerFunc(h.Login)))

	r.Handle("GET /companies", h.loggingMiddleware(http.HandlerFunc(h.GetCompanies)))
	r.Handle("GET /companies/{id}", h.loggingMiddleware(http.HandlerFunc(h.GetCompany)))
	r.Handle("POST /companies", h.loggingMiddleware(auth.JWTMiddleware(http.HandlerFunc(h.CreateCompany))))
	r.Handle("PATCH /companies/{id}", h.loggingMiddleware(auth.JWTMiddleware(http.HandlerFunc(h.UpdateCompany))))
	r.Handle("DELETE /companies/{id}", h.loggingMiddleware(auth.JWTMiddleware(http.HandlerFunc(h.DeleteCompany))))

	h.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.config.Server.Port),
		Handler: r,
	}

	go func() {
		h.log.Printf("HTTP server starting at '%d'", h.config.Server.Port)
		h.log.Println(h.httpServer.Addr)
		if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.log.Fatalf("Error on HTTP listen and serve: %v", err)
		}
	}()

	h.cdcChecker.Start()
}

func (h *Handler) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.httpServer.Shutdown(ctx); err != nil {
		h.log.Printf("Error on HTTP server shutdown: %v", err)
	} else {
		h.log.Print("HTTP server stopped gracefully")
	}

	h.cdcChecker.Stop()
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			h.log.Printf("Hit %s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		},
	)
}
