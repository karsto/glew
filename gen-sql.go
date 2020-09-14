package glew

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

type DBTypeCtx struct {
	Name       string
	Type       string
	Default    string
	IsPK       bool
	IsNullable bool
}

type DB struct {
}

// GenerateSQL - generates starter insert, read, list, update, delete, table create, table down sql statements.
func (_ *DB) GenerateSQL(ctx SQLCtx) (SQLStrings, error) {
	out := SQLStrings{}

	listF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\t\t%v,\n", cur)
	}
	insertColList := AggStrList(ctx.InsertFields, listF)
	insertColList = strings.Trim(insertColList, ",\n")
	listVal := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\t$%v,\n", cur)
	}
	insertValListStr := AggStrList(ctx.InsertFields, listVal)
	insertValListStr = strings.Trim(insertValListStr, ",\n")
	insertCtx := map[string]string{
		"TableName":        ctx.TableName,
		"InsertColList":    insertColList,
		"InsertValListStr": insertValListStr,
		"IDColName":        ctx.IDColName,
	}
	const insertTmpl = `
	INSERT INTO {{.TableName}} (
{{.InsertColList}}
	VALUES(
{{.InsertValListStr}}
	)
	RETURNING {{.IDColName}}
	`
	insertSQL, err := ExecuteTemplate("insertSQL", insertTmpl, insertCtx)
	if err != nil {
		return out, err
	}
	out.Insert = insertSQL

	readColList := AggStrList(ctx.DBFields, listF)
	readColList = strings.Trim(readColList, ",\n")
	listColCtx := map[string]string{
		"TableName":   ctx.TableName,
		"ReadColList": readColList,
	}
	const listTmpl = `
		SELECT
{{.ReadColList}}
		FROM {{.TableName}}
		`
	listSQL, err := ExecuteTemplate("listSQL", listTmpl, listColCtx)
	if err != nil {
		return out, err
	}
	out.List = listSQL

	readCtx := map[string]string{
		"TableName":   ctx.TableName,
		"ReadColList": readColList,
		"IDColName":   ctx.IDColName,
	}
	const readTmpl = `
		SELECT
{{.ReadColList}}
		FROM  {{.TableName}} WHERE tenant_id = $1 AND {{.IDColName}} = $2`
	readSQL, err := ExecuteTemplate("readSQL", readTmpl, readCtx)
	if err != nil {
		return out, err
	}
	out.Read = readSQL

	updateF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\t%v = $%v,\n", cur, idx+3)
	}
	putColList := AggStrList(ctx.PutFields, updateF)
	putColList = strings.Trim(putColList, ",\n")
	putCtx := map[string]string{
		"TableName":  ctx.TableName,
		"PutColList": putColList,
		"IDColName":  ctx.IDColName,
	}
	const putTmpl = `
	UPDATE {{.TableName}} SET
{{.PutColList}}
	WHERE tenant_id = $1 AND {{.IDColName}} = $2
	`
	putSQL, err := ExecuteTemplate("putSQL", putTmpl, putCtx)
	if err != nil {
		return out, err
	}
	out.Put = putSQL

	const deleteTmpl = `
	DELETE FROM {{.TableName}} WHERE tenant_id = ? AND {{.IDColName}} IN (?)
	`
	deleteSQL, err := ExecuteTemplate("deleteSQL", deleteTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.Delete = deleteSQL

	createColList := AggStrList(ctx.CreateStatements, listF)
	createColList = strings.Trim(createColList, ",\n")
	createCtx := map[string]string{
		"TableName":     ctx.TableName,
		"CreateColList": createColList,
	}
	const createTblTmpl = `
	CREATE TABLE {{.TableName}} (
{{.CreateColList}}
	);`
	createTblSQL, err := ExecuteTemplate("createTblSQL", createTblTmpl, createCtx)
	if err != nil {
		return out, err
	}
	out.CreateTable = createTblSQL

	const dropTblTmpl = `DROP TABLE {{.TableName}};`
	dropTblSQL, err := ExecuteTemplate("dropTblSQL", dropTblTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.DropTable = dropTblSQL

	return out, err
}

// GenerateCreateStatement - generates a field create statement for a sql table create script
func (_ *DB) GenerateCreateStatement(t DBTypeCtx) string {
	out := strings.Builder{}
	out.WriteString(t.Name)
	out.WriteString(" ")
	out.WriteString(t.Type)
	out.WriteString(" ")
	if !t.IsNullable {
		out.WriteString("NOT NULL ")
	}
	if t.IsPK {
		out.WriteString("PRIMARY KEY ")
	}
	if len(t.Default) > 0 {
		out.WriteString("DEFAULT ")
		out.WriteString(t.Default)
	}
	return out.String()
}

// GetSQLType - maps golang data type to corresponding postgres sql data type to be used in a create statement.
func (_ *DB) GetSQLType(t reflect.Type) string {
	out := "TODO"
	switch t.Kind() {
	case reflect.Int8, reflect.Uint8, reflect.Int16:
		out = "smallint"
	case reflect.Uint16, reflect.Int32:
		out = "integer"
	case reflect.Uint32, reflect.Int64, reflect.Int:
		out = "bigint"
	case reflect.Uint, reflect.Uint64:
		out = "bigint"
	case reflect.Float32:
		out = "real"
	case reflect.Float64:
		out = "precision"
	case reflect.Bool:
		out = "boolean"
	case reflect.String:
		out = "text"
		// TODO:
	// case  reflect. []byte:
	// 	out = "bytea"
	case reflect.Struct, reflect.Array, reflect.Map:
		out = "jsonb"
		// TODO:
		// case time.Time:
		// 	out = "timestamptz"
		// case net.IP:
		// 	out = "inet"
		// case net.IPNet:
		// 	out = ""
	}
	return out
}

// NewDBTypeCtx - looks at the struct tags - `db`, and `db2` to get additional information or overrides of sql specific field flags.
// `db` Tag - specificy the column name. Hijaking db struct tag from pgx.
// `db2` Tag - custom tag from glew
// db2:type - specify or override the db column type.
// db2:default - specify the db default value if any.
// db2:pk - set to enable primary key statement
// db2:notnull - set to enable not null statement.

func (db *DB) NewDBTypeCtx(t GoType) DBTypeCtx {
	name := t.Name
	if v, found := t.Tags.Lookup("db"); found {
		// TODO: nesting tags
		name = v
	}
	oType := db.GetSQLType(t.Type)
	if v, found := t.Tags.Lookup("db2:type"); found {
		// TODO: nesting tags
		oType = v
	}
	defaultVal := ""
	if v, found := t.Tags.Lookup("db2:default"); found {
		// TODO: nesting tags
		defaultVal = v
	}
	isPK := false
	if _, found := t.Tags.Lookup("db2:pk"); found {
		// TODO: nesting tags
		isPK = true
	}
	isNullable := false
	if _, found := t.Tags.Lookup("db2:notnull"); found {
		// TODO: nesting tags
		isNullable = true
	}
	out := DBTypeCtx{
		Name:       name,
		Type:       oType,
		Default:    defaultVal,
		IsPK:       isPK,
		IsNullable: isNullable,
	}
	return out
}

// NewSQLCtx - takes in the metadata for a given vertical and creates all related sql fields.
func (db *DB) NewSQLCtx(vertical VerticalMeta) SQLCtx {
	dbFields := []string{}
	createStatements := []string{}
	idColName := ".TODOidColName"
	for _, v := range vertical.Model.Fields {
		dbCtx := db.NewDBTypeCtx(v)
		if dbCtx.IsPK {
			idColName = dbCtx.Name
		}
		crtStmt := db.GenerateCreateStatement(dbCtx)
		createStatements = append(createStatements, crtStmt)
		dbFields = append(dbFields, dbCtx.Name)
	}

	insertFields := []string{}
	for _, v := range vertical.CreateModel.Fields {
		dbCtx := db.NewDBTypeCtx(v)
		insertFields = append(insertFields, dbCtx.Name)
	}

	putFields := []string{}
	for _, v := range vertical.CreateModel.Fields {
		dbCtx := db.NewDBTypeCtx(v)
		putFields = append(putFields, dbCtx.Name)
	}

	tableName := strcase.ToSnake(vertical.Name)
	out := SQLCtx{
		CreateStatements: createStatements,
		TableName:        tableName,
		DBFields:         dbFields,
		IDColName:        idColName,
		InsertFields:     insertFields,
		PutFields:        putFields,
	}
	return out
}
