package glew


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


