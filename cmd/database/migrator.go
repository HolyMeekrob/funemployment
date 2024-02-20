package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/holymeekrob/funemployment/internal/config"
	_ "modernc.org/sqlite"
	"os"
	"path"
	"strings"
	"time"
)

const migrationsDir = "cmd/database/migrations"

type migration struct {
	name     string
	filename string
	sql      []byte
}

func migrationName(filename string) string {
	parts := strings.Split(filename, "_")
	return parts[len(parts)-1]
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS __migrations
(
	id INT PRIMARY KEY NOT NULL,
	name TEXT NOT NULL,
	filename TEXT NOT NULL,
	timestamp TEXT NOT NULL
)`)
	return err
}

func existingMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM __migrations ORDER BY timestamp`)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func allMigrations(entries []os.DirEntry) ([]migration, error) {
	migrations := make([]migration, 0)
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}

		filename := entry.Name()
		migrationSql, err := os.ReadFile(path.Join(migrationsDir, filename))
		if err != nil {
			return migrations, err
		}

		migrations = append(migrations, migration{
			migrationName(filename),
			entry.Name(),
			migrationSql})
	}
	return migrations, nil
}

func skipAppliedMigrations(migrations []migration, appliedMigrations []string) ([]migration, error) {
	for i, appliedMigration := range appliedMigrations {
		if migrations[i].name != appliedMigration {
			return []migration{}, fmt.Errorf("unexpected migration: %s", appliedMigration)
		}
	}

	return migrations[len(appliedMigrations):], nil
}

func applyMigrations(migrations []migration, db *sql.DB) error {
	const newline = "\n"
	for _, migration := range migrations {
		now := time.Now().UTC().Format(time.RFC3339)
		recordMigration := fmt.Sprintf(`INSERT INTO __migrations (name, filename, timestamp) VALUES (%s, %s, %s);`, migration.name, migration.filename, now)
		fullMigration := fmt.Sprint("BEGIN TRANSACTION;", newline, migration.sql, newline, recordMigration, newline, "COMMIT;")

		_, err := db.Exec(fullMigration)
		if err != nil {
			return err
		}
	}

	return nil
}

func outputMigrations(migrations []migration) {
	const newline = "\n"
	if len(migrations) == 0 {
		println("No migrations to apply")
		return
	}

	lineBreaks := strings.Repeat(newline, 2)
	output := string(migrations[0].sql)
	for _, migration := range migrations[1:] {
		output = fmt.Sprint(output, lineBreaks, migration.sql)
	}

	println(output)
}

func main() {
	env := flag.String("env", "dev", "The name of the environment")
	cfg, err := config.Load(*env)
	if err != nil {
		panic(err.Error())
	}

	db, err := sql.Open("sqlite", cfg.Db)
	if err != nil {
		panic(err.Error())
	}

	err = ensureMigrationsTable(db)
	if err != nil {
		panic(err.Error())
	}

	appliedMigrations, err := existingMigrations(db)
	if err != nil {
		panic(err.Error())
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		panic(err.Error())
	}

	migrations, err := allMigrations(entries)
	if err != nil {
		panic(err.Error())
	}

	migrations, err = skipAppliedMigrations(migrations, appliedMigrations)
	if err != nil {
		panic(err.Error())
	}

	apply := flag.Bool("apply", false, "Will apply the changes to the database when set to true; otherwise pipes the scripts to standard output. Defaults to false.")

	if *apply {
		err = applyMigrations(migrations, db)
		if err != nil {
			panic(err.Error())
		}
	} else {
		outputMigrations(migrations)
	}
}
