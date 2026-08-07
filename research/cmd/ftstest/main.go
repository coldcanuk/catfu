package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	stmts := []string{
		`CREATE TABLE videos (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			upload_date TEXT
		);`,
		`CREATE VIRTUAL TABLE videos_fts USING fts5(
			title, description,
			content='videos', content_rowid='rowid',
			tokenize='porter unicode61 remove_diacritics 2'
		);`,
		`CREATE TRIGGER videos_ai AFTER INSERT ON videos BEGIN
			INSERT INTO videos_fts(rowid, title, description)
			VALUES (new.rowid, new.title, new.description);
		END;`,
		`INSERT INTO videos(id,title,description,upload_date) VALUES
			('a','Hello World','first video','2024-01-01'),
			('b','Goodbye Moon','second video','2024-06-01');`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			fmt.Println("ERR", err)
			os.Exit(1)
		}
	}
	rows, err := db.Query(`SELECT v.id, v.title FROM videos_fts f JOIN videos v ON v.rowid=f.rowid WHERE videos_fts MATCH 'hello' AND v.upload_date >= '2024-01-01'`)
	if err != nil {
		fmt.Println("query", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var id, title string
		_ = rows.Scan(&id, &title)
		fmt.Println("hit", id, title)
	}
	fmt.Println("FTS5 prototype OK")
}
