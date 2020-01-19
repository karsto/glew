package glew

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

type SField struct {
	Name string
	Type string
	Tags string
}

func GenerateStruct(structName string, fields []SField) (string, error) {
	structTpl := `
	type {{.StructName}} struct {
		{{.FieldsStr}}
	}
	`

	fields2 := []string{}
	for _, f := range fields {
		fields2 = append(fields2, fmt.Sprintf("%v %v `%v`", f.Name, f.Type, f.Tags))
	}

	listF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\t%v\n", cur)
	}
	fieldStr := AggStrList(fields2, listF)
	fieldStr = strings.Trim(fieldStr, "\n")
	ctx := map[string]string{
		"StructName": structName,
		"FieldsStr":  fieldStr,
	}
	content, err := ExecuteTemplate("structTpl", structTpl, ctx)
	if err != nil {
		return "", err
	}

	return content, nil
}

type BaseAPPCTX struct {
	ImportPath string
}

func GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
	files, err := ReadFiles("static", destDir)
	if err != nil {
		return files, err
	}

	// for _, v := range files {
	// v.Destination = strings.Trim(v.Destination, "static")
	// }

	basics, err := ReadFiles("templates/basic", destDir)
	if err != nil {
		return files, err
	}
	// for _, v := range basics {
	// v.Destination = strings.Trim(v.Destination, "templates/basic")
	// }

	basicCTX := map[string]string{
		"AppName":                      appName,
		"TODOProjectImportPath":        ctx.ImportPath,
		"TODODockerRegistry":           "{{.TODODockerRegistry}}",
		"TODODockerRepo":               "{{.TODODockerRepo}}",
		"TODOControllersRegistration":  "{{.TODOControllersRegistration}}",
		"TODOControllersRegistration2": "{{.TODOControllersRegistration2}}",
	}

	for i, v := range basics {
		content, err := ExecuteTemplate(fmt.Sprintf("bsc: %v", i)+destDir+v.FileName, v.Content, basicCTX)
		if err != nil {
			return files, err
		}
		basics[i].Content = content
	}
	files = append(files, basics...)

	return files, err
}

type ModelCtx struct {
	Model       string
	CreateModel string
	UpdateModel string
	Utilities   string
}

func GeneratePage(structName string) (string, error) {
	fields := []SField{
		SField{
			Name: "Records",
			Type: fmt.Sprintf("[]%v", structName),
			Tags: "json:\"records\"",
		},
		SField{
			Name: "Page",
			Type: "types.PagingInfo",
			Tags: "json:\"page\"",
		},
	}
	return GenerateStruct(structName+"Page", fields)
}

func GenerateModel(destDir, verticalName string, ctx ModelCtx) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	modelDest := path.Join(destDir, NewPaths().Model)
	fileName := fmt.Sprintf("%v.go", modelName)

	modelTpl := `package model
import (
	"time"
	"github.com/karsto/glew/common/types"
)

{{.Model}}

{{.CreateModel}}

{{.UpdateModel}}

{{.Utilities}}
`
	content, err := ExecuteTemplate("modelFunc", modelTpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: modelDest,
		FileName:    fileName,
	}

	return out, nil
}

type StoreCtx struct {
	TODOProjectImportPath    string
	TableName                string
	ModelNameTitleCase       string
	ModelNamePluralTitleCase string
	CreatePropertiesList     string
	UpdatePropertiesList     string
	SQL                      SQLStrings
}

func GenerateStoreFile(destDir, verticalName string, ctx StoreCtx) (FileContainer, error) {
	storeName := strcase.ToSnake(verticalName)
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)

	b, err := ioutil.ReadFile("templates/example-dal.go")
	if err != nil {
		return FileContainer{}, err
	}
	storeTmpl := string(b)

	content, err := ExecuteTemplate("storeFunc", storeTmpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: storeDest,
		FileName:    fileName,
	}
	return out, nil
}

var pluralizer = pluralize.NewClient() // save the bitsy

func NewControllerCtx(verticalName string, baseCTX BaseAPPCTX) ControllerCtx {
	pluralName := pluralizer.Plural(verticalName)
	out := ControllerCtx{
		ModelNameTitleCase:       strcase.ToCamel(verticalName),
		ModelNamePlural:          pluralName,
		ModelNamePluralTitleCase: strcase.ToCamel(pluralName),
		ModelNameDocs:            strcase.ToDelimited(verticalName, ' '),
		ModelIdFieldName:         "id",
		Route:                    strcase.ToKebab(verticalName),
		TODOProjectImportPath:    baseCTX.ImportPath,
	}
	return out
}

type ControllerCtx struct {
	ModelNameTitleCase       string
	ModelNamePlural          string
	ModelNamePluralTitleCase string
	ModelNameDocs            string
	ModelIdFieldName         string
	Route                    string
	TODOProjectImportPath    string
}

// returns Controller name, content, error
func GenerateControllerFile(destDir, verticalName string, ctx ControllerCtx) (FileContainer, error) {
	constrollerDest := path.Join(destDir, NewPaths().Controllers)
	name := strcase.ToSnake(verticalName)
	fileName := fmt.Sprintf("%v.go", name)

	b, err := ioutil.ReadFile("templates/example-controller.go")
	if err != nil {
		return FileContainer{}, err
	}
	controllerTmpl := string(b)
	content, err := ExecuteTemplate("controller", controllerTmpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: constrollerDest,
		FileName:    fileName,
	}

	return out, nil
}

type FeatureConfig struct {
	CopyBase    bool
	Store       bool
	Migrations  bool
	Controllers bool
	Models      bool
	Utilities   bool
}

func NewConfig() FeatureConfig {
	return FeatureConfig{
		CopyBase:    true,
		Store:       true,
		Migrations:  true,
		Controllers: true,
		Models:      true,
	}
}

func GenerateTrim(structName string, stringFieldNames []string) (string, error) {
	const trimTmpl = `
	func trim{{.StructName}}(m model.{{.StructName}}) model.{{.StructName}}{{print "{"}}{{ range  $value := .StringFieldNames }}
		m.{{$value}} = strings.TrimSpace(m.{{$value}}){{ end }}
		return m
	}`
	ctx := map[string]interface{}{
		"StructName":       structName,
		"StringFieldNames": stringFieldNames,
	}
	trimUtil, err := ExecuteTemplate("trim", trimTmpl, ctx)
	if err != nil {
		return "", err
	}
	return trimUtil, nil
}

func GenerateMapFunc(structName, targetName string, fields []string) (string, error) {
	// toMapPl := `
	// `
	toMapPl := `
	func (m *{{.StructName}}) To{{.TargetName}}() model.{{.TargetName}} {
		out := model.{{.TargetName}}{{print "{}"}}
		{{.MapStatement}}
		return out
	}
	`
	listf := func(idx int, cur, res string) string {
		out := fmt.Sprintf("m.%v = out.%v\n", cur, cur)
		return out
	}
	mapStmt := AggStrList(fields, listf)
	mapStmt = strings.Trim(mapStmt, "\n")
	ctx := map[string]interface{}{
		"StructName":   structName,
		"TargetName":   targetName,
		"MapStatement": mapStmt,
	}
	initFunc, err := ExecuteTemplate("toMapPl", toMapPl, ctx)
	if err != nil {
		return "", err
	}
	return initFunc, nil

}

// nilstatements[fieldname]newStatement
func GenerateInit(structName string, nilStatements map[string]string) (string, error) {
	const nilTmpl = `
	func initialize{{.StructName}}(m model.{{.StructName}}) model.{{.StructName}}{{"{"}}{{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
		return m
	}`
	ctx := map[string]interface{}{
		"StructName":    structName,
		"NilStatements": nilStatements,
	}
	initFunc, err := ExecuteTemplate("niltpl", nilTmpl, ctx)
	if err != nil {
		return "", err
	}
	return initFunc, nil
}

type SQLCtx struct {
	CreateStatements []string
	TableName        string
	InsertFields     []string
	PutFields        []string
	DBFields         []string
	IDColName        string
}

// aggFun index,vString, resultString -> additionToResultIfAny
func AggStrList(strs []string, aggFunc func(int, string, string) string) string {
	out := strings.Builder{}
	for i, v := range strs {
		agg := aggFunc(i, v, out.String())
		out.WriteString(agg)
	}
	return out.String()
}

func GenerateSQL(ctx SQLCtx) (SQLStrings, error) {
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

func GenerateMigrationFiles(destDir, verticalName string, sql SQLStrings, dbScriptIdxStart int) ([]FileContainer, error) {
	migrationsDest := path.Join(destDir, NewPaths().Core)
	verticalName = strcase.ToSnake(verticalName)

	out := []FileContainer{}
	fileName := fmt.Sprintf("%v_%s.up.sql", dbScriptIdxStart, verticalName)
	out = append(out, FileContainer{
		Content:     sql.CreateTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	fileName = fmt.Sprintf("%v_%s.drop.sql", dbScriptIdxStart, verticalName)
	out = append(out, FileContainer{
		Content:     sql.DropTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	return out, nil
}

type DBTypeCtx struct {
	Name       string
	Type       string
	Default    string
	IsPK       bool
	IsNullable bool
}

func GetCommon(left, right []GoType) []string {
	common := map[string]bool{}
	for _, v := range left {
		common[v.Name] = false
	}

	for _, v := range right {
		if _, found := common[v.Name]; found {
			common[v.Name] = true
		} else {
			common[v.Name] = false
		}
	}

	out := []string{}
	for k, v := range common {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func GenerateCreateStatement(t DBTypeCtx) string {
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
