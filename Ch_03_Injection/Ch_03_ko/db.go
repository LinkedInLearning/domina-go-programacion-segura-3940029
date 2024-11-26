package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const dbURL = "postgres://gopher:gopher@localhost:5432/pokemon"

var db *sql.DB

func init() {
	var err error

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
}

func updatePokedex(ctx context.Context, trainerID int, count string) error {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("pgx connect: %v", err)
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf("UPDATE pokedex SET maxPokemon = '%s' WHERE trainerId = %d", count, trainerID)

	_, err = conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("db exec: %v", err)
	}

	return nil
}
