package main

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"
)

type SQLStrings struct {
	Insert      string
	Read        string
	List        string
	Put         string
	Delete      string
	CreateTable string
	DropTable   string
}
type UtilityStrings struct {
	InitFunc string
	TrimFunc string
}

type ModelVertical struct {
	SQL       SQLStrings
	Utilities UtilityStrings
}

// TODO: GenerateApp(Verticals)
// createM and putM are optional
func GenerateVertical(model, createM, putM interface{}) (ModelVertical, error) {
	if createM == nil {
		createM = model
	}
	if putM == nil && createM != nil {
		putM = createM
	}
	if putM == nil {
		putM = model
	}

	out := ModelVertical{}

	name, fields, tags := getReflect(model)
	// TODO: idField
	idField := ""
	dbFields, nilFields, stringFields := getFieldTypes(fields, tags)

	templateCache := map[string]*template.Template{}
	// base template
	ctx := map[string]interface{}{}
	ctx["tableName"] = name // snake case lower

	// init base ctx, call all other features,
	// generate code
	return out, nil
}

// returns dbFields, nilFields, stringFields
func getFieldTypes(fields []reflect.StructField, tags map[string]reflect.StructTag) ([]reflect.StructField, []reflect.StructField, []reflect.StructField) {
	dbFields := []reflect.StructField{}
	nilFields := []reflect.StructField{}
	stringFields := []reflect.StructField{}
	for _, f := range fields {
		if v, found := tags[f.Name]; found && v.Get("db") != "" {
			dbFields = append(dbFields, f)
		}
		// //TODO: assume if type is not primitive then its nilable and print if structure
		if f.Type.String() == "string" {
			nilFields = append(nilFields, f)
		}
		if f.Type.String() == "string" {
			stringFields = append(stringFields, f)
		}
	}
	return dbFields, nilFields, stringFields

}

// returns m.somestringprop = strings.TrimSpace(m.somestringProp)
func getUtilities(modelName string, stringFields, nilFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (UtilityStrings, error) {
	out := UtilityStrings{}

	ctx["stringFields"] = stringFields
	const trimTmpl = `
				func trim{{{.modelName}}}(m {{{.modelName}}}) {{{.modelName}}}{
				{{ range _, $value := .stringFields }}
					m.{{.$value.Name}} = strings.TrimSpace(m.{{.$value.Name}})
				{{ end }}
					return m
				}`

	trimUtil, err := executeTemplate("trim", trimTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.TrimFunc = trimUtil

	ctx["nilFields"] = nilFields
	const nilTmpl = `
	func initialize{{.modelName}}(m {{.modelName}}) {{.modelName}}{
		{{ range _, $value := .nilFields }}
		if m.{{.$value.Name}} == nil {
			m.{{.$value.Name}} = {{.$value.Type.String()}}{{"{}"}}
		}
		{{ end }}
		return m
		}`
	initFunc, err := executeTemplate("nil", nilTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.InitFunc = initFunc
	return out, nil
}

// TODO: handle different model fields for update/create
func getSQL(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (SQLStrings, error) {
	out := SQLStrings{}
	const insertTmpl = `
	INSERT INTO {{.tableName}} (
				{{ range _, $value := .dbFields }}
				{{$value.Name},
				{{end}}
				VALUES({{range $idx, _ := .dbFields}}${{$.idx}}{{end}})
				RETURNING
				{{.idFieldName}}
	`
	insertSQL, err := executeTemplate("insertSQL", insertTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Insert = insertSQL

	const listTmpl = `
		SELECT
			{{ range _, $value := .dbFields }}
			{{$value.Name},
			{{end}}
		FROM {{.tableName}}
		`
	listSQL, err := executeTemplate("listSQL", listTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.List = listSQL

	const readTmpl = `
	SELECT
		{{ range _, $value := .dbFields }}
		{{$value.Name},
		{{end}}
	FROM  {{.tableName}} WHERE tenant_id = $1 AND {{.idFieldName}} = $2`
	readSQL, err := executeTemplate("readSQL", readTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Read = readSQL

	const putTmpl = `
	UPDATE {{.tableName}} SET
	{{ range idx, $value := .dbFields }}
	{{$value.Name} = ${{$idx+2}},
	{{end}}
	WHERE tenant_id = $1 AND {{.idFieldName}} = $2
	`
	putSQL, err := executeTemplate("putSQL", putTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Put = putSQL

	const deleteTmpl = `
	DELETE FROM {{.tableName}} WHERE tenant_id = ? AND {{.idFieldName}} IN (?)
	`
	deleteSQL, err := executeTemplate("deleteSQL", deleteTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Delete = deleteSQL

	// TODO: add special handling for idField, update_at, created_at
	// TODO: map go type to db type
	const createTblTmpl = `
	CREATE TABLE {{.tableName}} (
		{{.idFieldName}} serial PRIMARY KEY,
		{{ range idx, $value := .dbFields }}
		{{$value.Name} {{$value.dbType}},
		{{end}}
	);`
	createTblSQL, err := executeTemplate("createTblSQL", deleteTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.CreateTable = createTblSQL

	const dropTblTmpl = `DROP TABLE {{.tableName}};`
	dropTblSQL, err := executeTemplate("dropTblSQL", dropTblTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.DropTable = dropTblSQL

	return out, err
}

func executeTemplate(name, templateBody string, ctx map[string]interface{}, templateCache map[string]*template.Template) (string, error) {
	err := initIfNotFound(name, templateBody, templateCache)
	if err != nil {
		return "", err
	}

	tmpl, ok := templateCache[name]
	if !ok {
		return "", fmt.Errorf("template '%s' missing from cache", name)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func initIfNotFound(name, templateBody string, templateCache map[string]*template.Template) error {
	_, found := templateCache[name]
	if !found {
		tmpl, err := template.New(name).Parse(templateBody)
		if err != nil {
			return err
		}
		templateCache[name] = tmpl
	}
	return nil
}

// returns model name, fields, and tag map
func getReflect(m interface{}) (string, []reflect.StructField, map[string]reflect.StructTag) {
	t := reflect.TypeOf(m)

	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i <= t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}
	return t.Name(), fields, tagMap
}
