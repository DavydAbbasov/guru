package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	migrationtool "guru/utils/db-migration-tool"
	pgsqlmigrator "guru/utils/db-migration-tool/pgsql"
)

func main() {
	dir := flag.String("dir", env("MIGRATIONS_DIR", "./migrations"), "migrations directory")
	steps := flag.Int("steps", envInt("MIGRATIONS_STEPS", 1), "down steps")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <up|down|down_all|status>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	action := migrationtool.Action(flag.Arg(0))

	provider, err := pgsqlmigrator.NewPgsqlMigrator(&pgsqlmigrator.Config{
		DB: &pgsqlmigrator.DBConfig{
			Host: env("DB_HOST", "localhost"),
			Port: envInt("DB_PORT", 5432),
			User: env("DB_USER", "postgres"),
			Pass: env("DB_PASS", ""),
			Name: env("DB_NAME", "postgres"),
		},
		MigrationsDir: *dir,
		Steps:         *steps,
	})
	if err != nil {
		log.Fatalf("init migrator: %v", err)
	}
	defer provider.Close()

	if err := migrationtool.NewDBMigrator(provider).Execute(action); err != nil {
		log.Fatalf("%s failed: %v", action, err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
