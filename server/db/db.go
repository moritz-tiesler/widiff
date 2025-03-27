package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Table interface {
	Create() string
	Insert() string
}

type DiffsTable struct{}

func (diffsTable *DiffsTable) Name() string {
	return "diffs"
}

func (diffTable *DiffsTable) Create() string {
	return `
	create table if not exists
		diff (
			id integer not null primary key autoincrement,
			FromID    integer  ,
			FromRevID integer  ,
			FromNS    integer  ,
			FromTitle text     ,
			ToID      integer  ,
			ToRevID   integer  ,
			ToNS      integer  ,
			ToTitle   text     ,
			Body      text     ,
		);

	delete from diffs;
	`
}

func (diffTable *DiffsTable) Insert(
	FromRevID int,
	FromNS int,
	FromTitle string,
	ToID int,
	ToRevID int,
	ToNS int,
	ToTitle string,
	Body string,
) string {
	return fmt.Sprintf(`
		insert into %s(
			id,
			FromID,
			FromRevID ,
			FromNS    ,
			FromTitle ,
			ToID      ,
			ToRevID   ,
			ToNS      ,
			ToTitle   ,
			Body
		)
		values( ?, ?, ?, ?, ?, ?, ?, ?)`,
		diffTable.Name(),
	)
}

type DB struct {
	diffsTable DiffsTable
	*sql.DB
}

func NewDb() (*DB, error) {
	sqlDb, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		return nil, err
	}
	return &DB{
		diffsTable: DiffsTable{},
		DB:         sqlDb,
	}, nil
}

func (db *DB) Init() error {
	var dt DiffsTable
	stmt := dt.Create()
	_, err := db.Exec(stmt)
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
		return err
	}
	return nil
}

func TestDb() {

	db, err := NewDb()
	db.Init()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("こんにちは世界%03d", i))
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(name)

	_, err = db.Exec("delete from foo")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		log.Fatal(err)
	}

	rows, err = db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
