package main

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/davecgh/go-spew/spew"
)

type Device struct {
	ID          int                    `db:"id" rql:"filter,sort" json:"id,omitempty" example:"5"`
	HardwareID  string                 `db:"hardware_id" rql:"filter,sort" json:"hardwareId,omitempty" example:"1-1-1-1-1-1"`
	DisplayName string                 `db:"display_name" rql:"filter,sort" json:"displayName,omitempty"  example:"some name"`
	Location    interface{}            `db:"location" json:"_,omitempty" example:""`        // TODO: move to own resource as this is state// TODO: filter via location
	Metadata    map[string]interface{} `db:"metadata" json:"metadata,omitempty" example:""` // TODO: filter via metadata
}

func main() {
	fmt.Println("test")
	ctx := map[string]interface{}{}

	t := reflect.TypeOf(Device{})
	modelName := t.Name()
	ctx[""] = modelName

	for i := 0; i <= t.NumField(); i++ {
		f := t.Field(i)
		fieldName := f.Name
		columnName := f.Tag.Get("db")
		fieldType := f.Type
	}

	fmt.Println(ctx)
	spew.Dump(ctx)
}

type simpleField struct {
	Name string
}

// returns m.somestringprop = strings.TrimSpace(m.somestringProp)
func getTrimFunc(m interface{}) (string, error) {
	ctx := map[string]interface{}{}
	t := reflect.TypeOf(m)
	modelName := t.Name()
	ctx["modelName"] = modelName

	fields := getFields(m)

	stringFields := []reflect.StructField{}
	for _, f := range fields {
		if f.Type.String() == "string" {
			stringFields = append(stringFields, f)
		}
	}

	ctx["stringFields"] = stringFields

	const trimTmpl = `
				func trim{{{.modelName}}}(m {{{.modelName}}}) {{{.modelName}}}{
				{{ range _, $value := .stringFields }}
					m.{{.$value.Name}} = strings.TrimSpace(m.{{.$value.Name}})
				{{ end }}
					return m
				}`
	tmpl, err := template.New("trim").Parse(trimTmpl)

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), err
}

func getFields(m interface{}) []reflect.StructField {
	out := []reflect.StructField{}
	t := reflect.TypeOf(m)

	for i := 0; i <= t.NumField(); i++ {
		out = append(out, t.Field(i))
	}
	return out
}

// return m.someNonPrimProp != nil { m.someNonPrimProp = type{}}
func getNilDefaults(m interface{}) (string, error) {
	ctx := map[string]interface{}{}
	t := reflect.TypeOf(m)
	modelName := t.Name()
	ctx["modelName"] = modelName

	fields := getFields(m)

	nilFields := []reflect.StructField{}
	for _, f := range fields {
		// assume if type is not primitive then its nilable and print if structure
		if f.Type.String() == "string" {
			nilFields = append(nilFields, f)
		}
	}

	ctx["nilFields"] = nilFields

	const nilTmpl = `
				func initialize{{.modelName}}(m {{.modelName}}) {{.modelName}}{
				{{ range _, $value := .nilFields }}
				if m.{{.$value.Name}} == nil {
					m.{{.$value.Name}} = {.$value.Type.String()}{{"{}"}}
				}
				{{ end }}
				return m
				}`
	tmpl, err := template.New("trim").Parse(nilTmpl)

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), err
}

// returns read, list, insert, update, delete boiler
func getSql(m interface{}) {

	var insertTmpl := `
	INSERT INTO {.tableName} (
				{{ range _, $value := .columnNames }}
				{}

				if m.{{.$value.Name}} == nil {
					m.{{.$value.Name}} = {.$value.Type.String()}{{}}
				}
				{{end}}
		tenant_id, folder_id, device_type_id, hardware_id, display_name, location, metadata, is_active) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`

}
