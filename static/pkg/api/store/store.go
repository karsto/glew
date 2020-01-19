package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx"

	"github.com/jmoiron/sqlx"
	"github.com/karsto/glew/common/sqlutil"

	// DB DRIVER
	"github.com/jackc/pgx/stdlib"
	_ "github.com/jackc/pgx/stdlib"
)

func InitDB(config pgx.ConnConfig, maxConnections int, timeout time.Duration) (*sqlx.DB, error) {
	driverConfig := stdlib.DriverConfig{
		ConnConfig:   config,
		AfterConnect: nil,
	}
	stdlib.RegisterDriverConfig(&driverConfig)
	db, err := sql.Open("pgx", driverConfig.ConnectionString(""))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)
	return sqlx.NewDb(db, "pgx"), nil
}

type Store struct {
	db *sqlx.DB
}

func NewStore(config pgx.ConnConfig) *Store {
	db, err := InitDB(config, 50, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	return &Store{
		db: db,
	}
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
