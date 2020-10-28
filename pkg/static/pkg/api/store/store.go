package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/karsto/common/sqlutil"
	// DB DRIVER
)

func InitDB(driverName, connStr string, maxConnections int, timeout time.Duration) (*sqlx.DB, error) {
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)
	return sqlx.NewDb(db, driverName), nil
}

type Store struct {
	db *sqlx.DB
}

func NewPGXtore(connString string) (*Store, error) {
	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	parsedConnStr := stdlib.RegisterConnConfig(config)

	db, err := InitDB("pgx", parsedConnStr, 50, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (store *Store) deleteModel(sql string, tenantID int, IDs []int) (bool, int, error) {
	q, args, err := sqlx.In(sql, tenantID, IDs)
	if err != nil {
		return false, 0, err
	}
	q = store.db.Rebind(q)

	res, err := store.db.Exec(q, args...)
	if err != nil {
		return false, 0, err
	}
	count, _ := res.RowsAffected()

	return count > 0, int(count), nil
}

func (store *Store) getCount(tenantID int, tableName, whereExp string, filterArgs []interface{}) (int, error) {
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)
	sql2 := sqlutil.FmtSQL(countSQL, whereExp, "", 0, 0)
	sql2 = store.db.Rebind(sql2)
	args2 := sqlutil.FmtSQLArgs(tenantID, 0, 0, filterArgs)

	row := store.db.QueryRow(sql2, args2...)
	total := 0
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (store *Store) GetMigrationVersion() (int, error) {
	row := store.db.QueryRow("SELECT version FROM schema_migrations")
	ver := 0
	err := row.Scan(&ver)
	if err != nil {
		return 0, err
	}
	return ver, nil
}
