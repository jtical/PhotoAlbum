// Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"photoalbum.joelical.net/internal/data"
	"photoalbum.joelical.net/internal/jsonlog"
	"photoalbum.joelical.net/internal/mailer"
)

// The application version number
const version = "1.0.0"

// create a configuration struct
type config struct {
	port int
	env  string // diffrent type of environment, development, production
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	//stores variables for flag
	limiter struct {
		rps     float64 //request per second
		burst   int
		enabled bool
	}

	//stores config setting for mail server
	smtp struct {
		host     string
		port     int
		username string //from mailtrap settings
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// Dependency Injectiion, so its availabe to the handlers.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	//read in the flags that will be used to populate our config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production")
	//db flag
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("PA_DB_DSN"), "PostgreSQL_DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	//flags for the rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	//these are flags for the mailer
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "82bc4ad6e5b9c3", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "8e2dd5f137919b", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "PhotoAlbum <no-reply@photoalbum.icaljoel.net>", "SMTP sender")

	//use the flag.Func() function to parse our trusted origin flag from a string to a slice of string
	flag.Func("cors-trusted-origins", "Trusted CORS origin (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)

		return nil
	})

	flag.Parse()
	//create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	//create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	//close connection pool
	defer db.Close()
	//log the successful connection pool
	logger.PrintInfo("database connection pool established", nil)
	//create a new instance of our application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	//call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

// openDB() function returns a *sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	//test the connection pool. create a context with a 5-second time out deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx) //verifies a connection to db is still alive
	if err != nil {
		return nil, err
	}
	return db, nil
}
