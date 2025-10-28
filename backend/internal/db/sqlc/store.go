package db

import (
	"database/sql"
)

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Querier {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
