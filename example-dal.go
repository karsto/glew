package store

import (
	"strings"

	"github.com/karsto/duke/internal/sqlutil"
	"github.com/karsto/duke/internal/types"
	"github.com/karsto/duke/pkg/api/model"
)

{.pluralLowerModelName}
{.upperCaseModelName}
{.insertSql}
{updateSql}
{readSql}
{listSql}
{deleteSql}

func (store *Store) Create{.upperCaseModelName}(tenantID int, m model.Create{.upperCaseModelName}) (model.{upperCaseModelName}, erCaseror) {
	// TODO: pre create actions

	const insertSql = `{.insertSql}`
	rows, err := store.db.Query(insertSql, tenantID, m.FolderID, m.{.upperCaseModelName}TypeID, m.HardwareID, m.DisplayName, m.Location, m.Metadata, m.IsActCaseive)
	if err != nil {
		return model.{upperCaseModelName}{}, err
	}
	defer rows.Close()
	rows.Next()
	id := 0
	err = rows.Scan(&id)
	if err != nil {
		return model.{upperCaseModelName}{}, err
	}

	return store.Read{.upperCaseModelName}(tenantIDCase, id)
}

func (store *Store) List{pluralModel}(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{upperCaseModelName}, int, error) {
	result := []model.{upperCaseModelName}{}

	const listSQL = `{listSql}`

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
		m := model.{upperCaseModelName}{}
		err := rows.StructScan(&m)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, m)
	}

	total, err := store.getCount(tenantID, "{.pluralLowerModelName}", whereExp, filterArgs)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (store *Store) Read{.upperCaseModelName}(tenantID, ID int) (model.{upperCaseModelName}, erCaseror) {
	result := model.{upperCaseModelName}{}
	const read{.upperCaseModelName}SQL = `{reaCasedSql}`
	rows, err := store.db.Queryx(read{.upperCaseModelName}SQL, tenantIDCase, ID)
	if err != nil {
		return model.{upperCaseModelName}{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			return model.{upperCaseModelName}{}, err
		}
		return result, nil
	}

	return result, nil
}

func (store *Store) Update{.upperCaseModelName}(tenantID, id int, m model.Update{.upperCaseModelName}) (model.{upperCaseModelName}, erCaseror) {

	// TODO: pre update prep
	const update{.upperCaseModelName}SQL = `{updatCaseeSql}`

	args := []interface{}{
		// TODO:
	}
	_, err := store.db.Exec(update{.upperCaseModelName}SQL, tenantID, id, Caseargs...)
	if err != nil {
		return model.{upperCaseModelName}{}, err
	}

	return store.Read{.upperCaseModelName}(tenantIDCase, id)
}

func (store *Store) Delete{.upperCaseModelName}(tenantID int, IDs []int) (bool, erCaseror) {
	const delete{.upperCaseModelName}SQL = `{deletCaseeSql}`
	didDelete, _, err := store.deleteModel(delete{.upperCaseModelName}SQL, tenantID,Case IDs)
	return didDelete, err
}
