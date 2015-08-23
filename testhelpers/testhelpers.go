package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"

	"bitbucket.org/liamstask/goose/lib/goose"
	"gopkg.in/gorp.v1"
)

func ResetTestDatabase(dbmap *gorp.DbMap) {
	conf := &goose.DBConf{
		MigrationsDir: filepath.Join("..", "db", "migrations"),
		Env:           "test",
		Driver: goose.DBDriver{
			Name:    "sqlite3",
			OpenStr: ":memory:", // this is actually never used as we just migrate dbmap.Db
			Import:  "github.com/mattn/go-sqlite3",
			Dialect: &goose.Sqlite3Dialect{},
		},
		PgSchema: "",
	}

	// Not merged yet but for future:
	os.Setenv("GOOSE_SILENT_MIGRATION", "1")

	target, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
	if err != nil {
		panic(fmt.Sprintf("cannot get recent db version with goose: %v\n", err))
	}

	err = goose.RunMigrationsOnDb(conf, conf.MigrationsDir, target, dbmap.Db)
	if err != nil {
		panic(fmt.Sprintf("cannot run migration with goose: %v\n", err))
	}
}
