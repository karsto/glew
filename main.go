package main

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	fmt.Println("test")

	ctx := map[string]interface{}{}

	// ctx["dbFields"] = dbFields
	// tableName := modelName // TODO: snake case struct name
	// ctx["tableName"] = tableName

	// ctx["idFieldName"] = idFieldName

	fmt.Println(ctx)
	spew.Dump(ctx)
}

type Device struct {
}

type ModelVertical struct {
}

func generateVertical(m interface{}) (ModelVertical, error) {
	out := ModelVertical{}
	name, fields, tags := getReflect(m)
	dbFields, nilFields, stringFields := getFieldTypes(fields, tags)

	ctx := map[string]interface{}
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
func getTrimFunc(modelName string, stringFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	ctx["stringFields"] = stringFields

	const trimTmpl = `
				func trim{{{.modelName}}}(m {{{.modelName}}}) {{{.modelName}}}{
				{{ range _, $value := .stringFields }}
					m.{{.$value.Name}} = strings.TrimSpace(m.{{.$value.Name}})
				{{ end }}
					return m
				}`

	execResult, err := executeTemplate("trim", trimTmpl, ctx, templateCache)
	return execResult, err
}

// return m.someNonPrimProp != nil { m.someNonPrimProp = type{}}
func getNilDefaults(modelName string, nilFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
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
	execResult, err := executeTemplate("nil", nilTmpl, ctx, templateCache)
	return execResult, err
}

// returns dbFields, nilableFields, stringFields

// returns read, list, insert, update, delete boiler
func getInsertSQL(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	const insertTmpl = `
	INSERT INTO {{.tableName}} (
				{{ range _, $value := .dbFields }}
				{{$value.Name},
				{{end}}
				VALUES({{range $idx, _ := .dbFields}}${{$.idx}}{{end}})
				RETURNING
				{{.idFieldName}}
	`
	execResult, err := executeTemplate("insertSql", insertTmpl, ctx, templateCache)
	return execResult, err
}

func getListSQL(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	const listSQL = `
	SELECT
		{{ range _, $value := .dbFields }}
		{{$value.Name},
		{{end}}
	FROM {{.tableName}}
	`
	execResult, err := executeTemplate("listSQL", listSQL, ctx, templateCache)
	return execResult, err
}
func getReadSQL(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	const readDeviceSQL = `
	SELECT
		{{ range _, $value := .dbFields }}
		{{$value.Name},
		{{end}}
	FROM  {{.tableName}} WHERE tenant_id = $1 AND {{.idFieldName}} = $2`
	execResult, err := executeTemplate("readDeviceSQL", readDeviceSQL, ctx, templateCache)
	return execResult, err
}

func getUpdatePutSql(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	const updateSql = `
	UPDATE {{.tableName}} SET
	{{ range idx, $value := .dbFields }}
	{{$value.Name} = ${{$idx+2}},
	{{end}}
	WHERE tenant_id = $1 AND {{.idFieldName}} = $2
	`
	execResult, err := executeTemplate("updateSql", updateSql, ctx, templateCache)
	return execResult, err
}

func getDeletSQL(modelName, idFieldName string, dbFields []reflect.StructField, templateCache map[string]*template.Template, ctx map[string]interface{}) (string, error) {
	const deleteSQL = `
	DELETE FROM {{.tableName}} WHERE tenant_id = ? AND {{.idFieldName}} IN (?)
	`
	execResult, err := executeTemplate("deleteSQL", deleteSQL, ctx, templateCache)
	return execResult, err
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
