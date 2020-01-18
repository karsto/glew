package store

import (
	"strings"

	"github.com/karsto/glew/internal/sqlutil"
)

func (store *Store) Create{{.ModelNameTitleCase}}(tenantID int, m model.Create{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	const insertSql = `{{.SQL.Insert}}`

	args := []interface{}{{"{"}}
{{.CreatePropertiesList}}
	}

	rows, err := store.db.Query(insertSql, tenantID, args...)
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

	return store.Read{{.ModelNameTitleCase}}(tenantID, id)
}

func (store *Store) List{{.ModelNamePluralTitleCase}}(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{{.ModelNameTitleCase}}, int, error) {
	result := []model.{{.ModelNameTitleCase}}{}

	const listSQL = `{{.SQL.List}}`

	whereExp := sqlutil.AddTenantCheck(tenantID, filterExp)
	sql := sqlutil.FmtSQL(listSQL, whereExp, sortExp, limit, offset)
	args := sqlutil.FmtSQLArgs(tenantID, limit, offset, filterArgs)
	sql = store.db.Rebind(sql)

	rows, err := store.db.Queryx(sql, args...)
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

	total, err := store.getCount(tenantID, "{{.TableName}}", whereExp, filterArgs)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (store *Store) Read{{.ModelNameTitleCase}}(tenantID, ID int) (model.{{.ModelNameTitleCase}}, error) {
	result := model.{{.ModelNameTitleCase}}{}
	const read{{.ModelNameTitleCase}}SQL = `{{.SQL.Read}}`
	rows, err := store.db.Queryx(read{{.ModelNameTitleCase}}SQL, tenantID, ID)
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

func (store *Store) Update{{.ModelNameTitleCase}}(tenantID, id int, m model.Update{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	const update{{.ModelNameTitleCase}}SQL = `{{.SQL.Put}}`

	args := []interface{}{{"{"}}
{{.UpdatePropertiesList}}
	}
	_, err := store.db.Exec(update{{.ModelNameTitleCase}}SQL, tenantID, id, args...)
	if err != nil {
		return model.{{.ModelNameTitleCase}}{}, err
	}

	return store.Read{{.ModelNameTitleCase}}(tenantID, id)
}

func (store *Store) Delete{{.ModelNameTitleCase}}(tenantID int, IDs []int) (bool, error) {
	const delete{{.ModelNameTitleCase}}SQL = `{{.SQL.Delete}}`
	didDelete, _, err := store.deleteModel(delete{{.ModelNameTitleCase}}SQL, tenantID, IDs)
	return didDelete, err
}