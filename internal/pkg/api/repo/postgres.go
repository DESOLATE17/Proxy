package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"proxy/internal/models"
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

func (r *Repo) SaveRequest(request models.Request) (int, error) {
	requestStr, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO request(request) VALUES ($1) RETURNING request_id;`

	var reqId int
	row := r.db.QueryRow(context.Background(), query, requestStr)
	err = row.Scan(&reqId)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return reqId, err
}

func (r *Repo) SaveResponse(reqId int, response models.Response) error {
	responseStr, err := json.Marshal(response)
	if err != nil {
		return err
	}

	query := `INSERT INTO response(response, request_id) VALUES ($1, $2) RETURNING response_id;`

	var responseId int
	row := r.db.QueryRow(context.Background(), query, responseStr, reqId)
	err = row.Scan(&responseId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return err
}

func (r *Repo) GetRequest(requestId int) (models.Request, error) {
	query := `SELECT request FROM request WHERE request_id = $1;`

	requestStr := ""
	row := r.db.QueryRow(context.Background(), query, requestId)
	err := row.Scan(&requestStr)
	if err != nil {
		fmt.Println(err)
		return models.Request{}, err
	}

	var request models.Request
	err = json.Unmarshal([]byte(requestStr), &request)
	if err != nil {
		fmt.Println(err)
		return models.Request{}, err
	}
	request.Id = requestId
	return request, err

}

func (r *Repo) AllRequests() ([]models.Request, error) {
	query := `SELECT request_id, request FROM request;`

	requests := make([]models.Request, 0)
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return requests, err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.Request
		reqId := 0
		requestStr := ""
		err = rows.Scan(&reqId, &requestStr)
		if err != nil {
			fmt.Println(err)
			return requests, err
		}

		err = json.Unmarshal([]byte(requestStr), &request)
		request.Id = reqId
		if err != nil {
			fmt.Println(err)
			return requests, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}
