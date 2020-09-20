package store

import (
	"strings"

	"{{.TODOProjectImportPath}}/pkg/api/model"
)

const readTenantSQL = `SELECT * FROM tenants WHERE id = $1`

func (store *Store) ReadTenant(ID int) (model.Tenant, error) {
	result := model.Tenant{}

	rows, err := store.db.Queryx(readTenantSQL, ID)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		err := rows.StructScan(&result)
		if err != nil {
			return result, err
		}
		return result, nil
	}

	return result, nil
}

const updateTenantSQL = `
UPDATE tenants SET
name = $2
,is_active = $3
,metadata = $4
,updated_at = NOW()
WHERE id = $1
`

func (store *Store) UpdateTenant(ID int, m model.UpdateTenant) (model.Tenant, error) {
	m.Name = strings.TrimSpace(m.Name)
	_, err := store.db.Exec(updateTenantSQL, ID, m.Name, m.IsActive, m.Metadata)
	if err != nil {
		return model.Tenant{}, err
	}
	return store.ReadTenant(ID)
}
