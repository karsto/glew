package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"github.com/karsto/common/sqlutil"
	store "github.com/karsto/glew/templates"
	//"{{.TODOProjectImportPath}}/pkg/api/model"
)

type IDB interface {
	GetDB() *sqlx.DB
 	DeleteModel(sql string, tenantID int, IDs []int) (bool, int, error)
 	GetCount(tenantID int, tableName, whereExp string, filterArgs []interface{}) (int, error)
 	GetMigrationVersion() (int, error)
}

// type Store struct {
// 	// TODO: loop all
// ICRUDModel
//}

type ICRUD{{.ModelNameTitleCase}} interface
{
	Create(tenantID int, m model.Create{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error)
	List(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{{.ModelNameTitleCase}}, int, error)
	Read(tenantID, ID int) (model.{{.ModelNameTitleCase}}, error)
	Update(tenantID, id int, m model.Update{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error)
	Delete(tenantID int, IDs []int) (bool, error)
}

func New{{.ModelNameTitleCase}}Store(db IDB) ICRUD{{.ModelNameTitleCase}}{
	out := {{.ModelNameTitleCase}}Store{
		db: db,
	}
	return out
}

type {{.ModelNameTitleCase}}Store struct {
	db IDB
}

func (store *{{.ModelNameTitleCase}}Store) Create(tenantID int, m model.Create{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	const insertSql = `{{.SQL.Insert}}`

	args := []interface{}{{"{"}}
		tenantID,
{{.CreatePropertiesList}}
	}

	rows, err := store.GetDB().Query(insertSql, args...)
	if err != nil {
		return model.{{.ModelNameTitleCase}}{}, err
	}
	defer rows.Close()
	rows.Next()
	id := 0
	err = rows.Scan(&id)
	if err != nil {
		return model.{{.ModelNameTitleCase}}{}, err
	}

	return store.Read(tenantID, id)
}

func (store *{{.ModelNameTitleCase}}Store) List(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{{.ModelNameTitleCase}}, int, error) {
	result := []model.{{.ModelNameTitleCase}}{}

	const listSQL = `{{.SQL.List}}`

	whereExp := sqlutil.AddTenantCheck(tenantID, filterExp)
	sql := sqlutil.FmtSQL(listSQL, whereExp, sortExp, limit, offset)
	args := sqlutil.FmtSQLArgs(tenantID, limit, offset, filterArgs)
	sql = store.GetDB().Rebind(sql)

	rows, err := store.GetDB().Queryx(sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		m := model.{{.ModelNameTitleCase}}{}
		err := rows.StructScan(&m)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, m)
	}

	total, err := store.GetDB().GetCount(tenantID, "{{.TableName}}", whereExp, filterArgs)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (store *{{.ModelNameTitleCase}}Store) Read(tenantID, ID int) (model.{{.ModelNameTitleCase}}, error) {
	result := model.{{.ModelNameTitleCase}}{}
	const read{{.ModelNameTitleCase}}SQL = `{{.SQL.Read}}`
	rows, err := store.GetDB().Queryx(read{{.ModelNameTitleCase}}SQL, tenantID, ID)
	if err != nil {
		return model.{{.ModelNameTitleCase}}{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			return model.{{.ModelNameTitleCase}}{}, err
		}
		return result, nil
	}

	return result, nil
}

func (store *{{.ModelNameTitleCase}}Store) Update(tenantID, id int, m model.Update{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	const update{{.ModelNameTitleCase}}SQL = `{{.SQL.Put}}`

	args := []interface{}{{"{"}}
		tenantID,
		id,
{{.UpdatePropertiesList}}
	}
	_, err := store.GetDB().Exec(update{{.ModelNameTitleCase}}SQL, args...)
	if err != nil {
		return model.{{.ModelNameTitleCase}}{}, err
	}

	return store.Read(tenantID, id)
}

func (store *{{.ModelNameTitleCase}}Store) Delete(tenantID int, IDs []int) (bool, error) {
	const delete{{.ModelNameTitleCase}}SQL = `{{.SQL.Delete}}`
	didDelete, _, err := store.GetDB().DeleteModel(delete{{.ModelNameTitleCase}}SQL, tenantID, IDs)
	return didDelete, err
}