package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewPostgresDB(str string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), str)
	if err != nil {
		log.Println("Error while connecting to DB", err)
		return nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Println("Error while ping to DB", err)
		return nil, err
	}
	log.Println("connected to postgres")
	return db, err
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}
