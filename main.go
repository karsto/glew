package main

import (
	"bytes"
	"fmt"
	"github.com/karst/glew/internal/files"
	"path"
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
	Reflect   SimpleReflect
	SQL       SQLStrings
	Utilities UtilityStrings
}

type SimpleReflect struct {
	Name         string
	IDFieldName  string
	Fields       []reflect.StructField
	Tags         map[string]reflect.StructTag
	DBFields     []reflect.StructField
	NilFields    []reflect.StructField
	StringFields []reflect.StructField
}

func main() {

}

func GetMeta(m interface{}) (SimpleReflect, error) {
	t := reflect.TypeOf(m)

	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i <= t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}

	dbFields := []reflect.StructField{}
	nilFields := []reflect.StructField{}
	stringFields := []reflect.StructField{}
	for _, f := range fields {
		if v, found := tagMap[f.Name]; found && v.Get("db") != "" {
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

	out := SimpleReflect{
		Name:         t.Name(),
		Fields:       fields,
		Tags:         tagMap,
		DBFields:     dbFields,
		NilFields:    nilFields,
		StringFields: stringFields,
	}
	return out, nil
}

// TODO: db migrations base project copy - 0 and tenant
// TODO: add tenant fields to base

type Paths struct {
	DB          string
	Migrations  string
	Core        string
	Controllers string
	Store       string
	Model       string
}

func NewPaths() Paths {
	// TODO: save the bits
	return Paths{
		DB:          "db",
		Migrations:  "db/migrations",
		Core:        "db/migrations/core",
		Controllers: "pkg/api/controllers",
		Store:       "pkg/api/store",
		Model:       "pkg/api/model",
	}
}

func WriteMigration(v ModelVertical, destDir string, dbScriptIdxStart int) error {
	// TODO:
	migrationsDest := path.Join(destDir, NewPaths().Migrations)
	name := v.Reflect.Name // TODO: underscore lower case
	fileName := fmt.Sprintf("%v_%s.up.sql", dbScriptIdxStart, name)
	err := files.WriteFile(migrationsDest, fileName, "TODO:")
	if err != nil {
		return err
	}
	fileName = fmt.Sprintf("%v_%s.drop.sql", dbScriptIdxStart, name)
	err = files.WriteFile(migrationsDest, fileName, "TODO:")
	if err != nil {
		return err
	}
	return nil
}

func WriteController(v ModelVertical, destDir string) error {
	controllerName := v.Reflect.Name // TODO: underscore lower
	constrollerDest := path.Join(destDir, NewPaths().Controllers)
	fileName := fmt.Sprintf("%v.go", controllerName)
	err := files.WriteFile(constrollerDest, fileName, "TODO:")
	if err != nil {
		return err
	}
	return nil
}

func WriteStore(v ModelVertical, destDir string) error {
	storeName := v.Reflect.Name // TODO: underscore lower
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)
	err := files.WriteFile(storeDest, fileName, "TODO:")
	if err != nil {
		return err
	}
	return nil
}

func WriteModel(v ModelVertical, destDir string) error {
	modelName := v.Reflect.Name // TODO: underscore lower
	modelDest := path.Join(destDir, NewPaths().Model)
	fileName := fmt.Sprintf("%v.go", modelName)
	err := files.WriteFile(modelDest, fileName, "TODO:")
	if err != nil {
		return err
	}
	return nil
}

type Config struct {
	CopyBase         bool
	WriteStore       bool
	WriteMigrations  bool
	WriteControllers bool
	WriteModels      bool
}

func NewConfig() Config {
	return Config{
		CopyBase:         true,
		WriteStore:       true,
		WriteMigrations:  true,
		WriteControllers: true,
		WriteModels:      true,
	}
}

func GenerateApp(verticals []ModelVertical, destDir string) error {
	// copy base
	err := files.CopyDirectory("./base-project", destDir)
	if err != nil {
		return err
	}
	cfg := NewConfig()
	for _, v := range verticals {
		if cfg.WriteStore {
			WriteStore(v, destDir)
		}
		if cfg.WriteMigrations {
			migrationStartId := 2 // TODO: read from directory automatically
			WriteMigration(v, destDir, migrationStartId)
		}
		if cfg.WriteControllers {
			WriteController(v, destDir)
		}
		if cfg.WriteModels {
			WriteModel(v, destDir)
		}
	}

	return nil
}

// createM and putM are optional
func GenerateVertical(model, createM, putM interface{}, templateCache map[string]*template.Template) (ModelVertical, error) {
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
	modelMeta, err := GetMeta(model)
	if err != nil {
		return out, err
	}

	createMeta, err := GetMeta(createM)
	if err != nil {
		return out, err
	}
	putMeta, err := GetMeta(putM)
	if err != nil {
		return out, err
	}

	ctx := map[string]interface{}{}
	ctx["dbFields"] = modelMeta.DBFields
	ctx["insertFields"] = createMeta.DBFields // insertFields
	ctx["putFields"] = putMeta.DBFields
	ctx["tableName"] = modelMeta.Name // TODO: snake case lower

	utilities, err := getUtilities(templateCache, ctx)
	SQLStrings, err := getSQL(templateCache, ctx)

	out.Reflect = modelMeta
	out.Utilities = utilities
	out.SQL = SQLStrings
	return out, nil
}

// returns m.somestringprop = strings.TrimSpace(m.somestringProp)
func getUtilities(templateCache map[string]*template.Template, ctx map[string]interface{}) (UtilityStrings, error) {
	out := UtilityStrings{}
	const trimTmpl = `
				func trim{{{.modelName}}}(m {{{.modelName}}}) {{{.modelName}}}{
				{{ range _, $value := .stringFields }}
					m.{{.$value.Name}} = strings.TrimSpace(m.{{.$value.Name}})
				{{ end }}
					return m
				}`

	trimUtil, err := ExecuteTemplate("trim", trimTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.TrimFunc = trimUtil

	const nilTmpl = `
	func initialize{{.modelName}}(m {{.modelName}}) {{.modelName}}{
		{{ range _, $value := .nilFields }}
		if m.{{.$value.Name}} == nil {
			m.{{.$value.Name}} = {{.$value.Type.String()}}{{"{}"}}
		}
		{{ end }}
		return m
		}`
	initFunc, err := ExecuteTemplate("nil", nilTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.InitFunc = initFunc
	return out, nil
}

// TODO: handle different model fields for update/create
func getSQL(templateCache map[string]*template.Template, ctx map[string]interface{}) (SQLStrings, error) {
	out := SQLStrings{}
	const insertTmpl = `
	INSERT INTO {{.tableName}} (
				{{ range _, $value := .insertFields }}
				{{$value.Name},
				{{end}}
				VALUES({{range $idx, _ := .insertFields}}${{$.idx}}{{end}})
				RETURNING
				{{.idFieldName}}
	`
	insertSQL, err := ExecuteTemplate("insertSQL", insertTmpl, ctx, templateCache)
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
	listSQL, err := ExecuteTemplate("listSQL", listTmpl, ctx, templateCache)
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
	readSQL, err := ExecuteTemplate("readSQL", readTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Read = readSQL

	const putTmpl = `
	UPDATE {{.tableName}} SET
	{{ range idx, $value := .putFields }}
	{{$value.Name} = ${{$idx+2}},
	{{end}}
	WHERE tenant_id = $1 AND {{.idFieldName}} = $2
	`
	putSQL, err := ExecuteTemplate("putSQL", putTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Put = putSQL

	const deleteTmpl = `
	DELETE FROM {{.tableName}} WHERE tenant_id = ? AND {{.idFieldName}} IN (?)
	`
	deleteSQL, err := ExecuteTemplate("deleteSQL", deleteTmpl, ctx, templateCache)
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
	createTblSQL, err := ExecuteTemplate("createTblSQL", deleteTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.CreateTable = createTblSQL

	const dropTblTmpl = `DROP TABLE {{.tableName}};`
	dropTblSQL, err := ExecuteTemplate("dropTblSQL", dropTblTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.DropTable = dropTblSQL

	return out, err
}

func ExecuteTemplate(name, templateBody string, ctx map[string]interface{}, templateCache map[string]*template.Template) (string, error) {
	err := InitIfNotFound(name, templateBody, templateCache)
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

func InitIfNotFound(name, templateBody string, templateCache map[string]*template.Template) error {
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
