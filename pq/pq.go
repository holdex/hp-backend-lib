package libpq

import (
	"database/sql"

	"github.com/holdex/hp-backend-lib/log"
)

func Open(dataSourceName string) *sql.DB {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		liblog.Fatalf("failed to connect to postgres [%s]: %v", dataSourceName, err)
	}
	return db
}
