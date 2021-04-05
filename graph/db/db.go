package db

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/vickywane/api/graph/model"
	"os"
)

func createSchema(db *pg.DB) error {
	for _, models := range []interface{}{(*model.User)(nil), (*model.User)(nil)}{
		if err := db.CreateTable(models, &orm.CreateTableOptions{
			IfNotExists: true, FKConstraints: false, // Todo: turned this off because of VOLUNTEER table. Check out later!!
		}); err != nil {
			panic(err)
		}
	}

	return nil
}

func Connect() *pg.DB {
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	DB_ADDR := os.Getenv("DB_ADDR")
	DB_USER := os.Getenv("DB_USER")

	connStr := fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v?sslmode=require",
		DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_NAME )

	opt, err := pg.ParseURL(connStr)

	if err != nil {
	    panic(err)
	}

	db := pg.Connect(opt)

	if schemaErr := createSchema(db); schemaErr != nil {
		panic(schemaErr)
	}

	if _, DBStatus := db.Exec("SELECT 1"); DBStatus != nil {
		panic("PostgreSQL is down")
	}

	return db
}