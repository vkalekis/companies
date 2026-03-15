package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vkalekis/companies/internal/cdc"
	"github.com/vkalekis/companies/internal/config"
	"github.com/vkalekis/companies/internal/handler"
	"github.com/vkalekis/companies/internal/migrations"
)

type App struct {
	config  *config.Config
	log     *log.Logger
	pool    *pgxpool.Pool
	handler *handler.Handler

	cdcOperator cdc.Operator
}

func NewApp(c *config.Config, log *log.Logger) (*App, error) {
	app := App{
		config: c,
		log:    log,
	}

	// First get the cdc operator from config
	var err error
	switch app.config.CDC.Operator {
	case config.CDCLog:
		app.cdcOperator = &cdc.LogCDCOperator{}
	case config.CDCKafka:
		app.cdcOperator, err = cdc.NewKafkaCDCOperator(c)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported cdc format")
	}

	return &app, nil
}

func (app *App) Start() error {

	// postgres://<username>:<password>@<host>:<port>/<dbname>
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		app.config.Database.Username,
		app.config.Database.Password,
		app.config.Database.Host,
		app.config.Database.Port,
		app.config.Database.DBName)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("error parsing dbURL %s: %v", dbURL, err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("error creating db pool for %s: %v", dbURL, err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("error pinging db %s: %v", dbURL, err)
	}

	app.pool = pool

	// !! Run migrations first !!
	if err := migrations.Run(pool); err != nil {
		return fmt.Errorf("error during migrations: %w", err)
	}

	// Start DB+HTTP handler
	app.handler = handler.New(app.config, pool, app.cdcOperator)
	app.handler.Start()

	return nil
}

func (app *App) Stop() {
	app.handler.Stop()

	if app.pool != nil {
		app.pool.Close()
	}
}
