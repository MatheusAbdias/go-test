package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"

	scs "github.com/alexedwards/scs/v2"

	"webapp/pkg/data"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"
)

type application struct {
	DSN     string
	DB      repository.DatabaseRepo
	Session *scs.SessionManager
}

func main() {
	gob.Register(data.User{})

	app := application{}

	flag.StringVar(
		&app.DSN,
		"dsn",
		"host=localhost port=5432 user=web_user password=password dbname=web sslmode=disable timezone=UTC connect_timeout=5",
		"Posgtres connection",
	)
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	// get a session manager
	app.Session = getSession()

	// print out a message
	log.Println("Starting server on port 8080...")

	// start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
