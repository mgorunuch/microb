package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/postgres"
)

func main() {
	core.Init()
	cleanup := postgres.Init(context.Background())
	defer cleanup()

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		core.Logger.Fatal("no input provided")
	}

	id := scanner.Text()

	var html string
	err := postgres.Pool.QueryRow(context.Background(), `
		select html 
		from chrome_visits 
		where id = $1
	`, id).Scan(&html)

	if err == pgx.ErrNoRows {
		core.Logger.Fatal("visit not found")
	}
	if err != nil {
		core.Logger.Fatal(err)
	}

	fmt.Print(html)
}
