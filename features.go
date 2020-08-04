package glew

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// TODO: Refactor and consolidate file generate pattern?

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

type StoreTemplateVueCtx struct {
	Resource string
}

func NewStoreTemplateVueCtx(vertical VerticalMeta) (StoreTemplateVueCtx, error) {
	out := StoreTemplateVueCtx{
		Resource: strcase.ToKebab(vertical.Name),
	}
	return out, nil
}

func GenerateStoreVueFile(destDir, verticalName string, ctx StoreTemplateVueCtx) (FileContainer, error) {
	modelName := strcase.ToLowerCamel(verticalName)
	uiStoreDest := path.Join(destDir, NewPaths().UIStore)
	fileName := fmt.Sprintf("%v.js", modelName)

	b, err := ioutil.ReadFile("templates/ui/store-template.js") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	storeTmpl := string(b)

	content, err := ExecuteTemplate("uiStore", storeTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: uiStoreDest,
		FileName:    fileName,
	}
	return out, nil
}

type NewTemplateVueCtx struct {
	ModelFieldsMeta          []ModelFieldMeta
	ResourceRoute            string
	ModelTitleCaseName       string
	TitleCaseModelName       string
	CamelCaseModelName       string
	CamelCasePluralModelName string
	TitleCaseModelPluralName string
	FormMapStatment          string
	FormDefaultStatement     string
}

func GenerateMapStatement(types []GoType) string {
	out := strings.Builder{}
	for _, v := range types {
		name := strcase.ToLowerCamel(v.Name)
		stmt := fmt.Sprintf("%s:'%s',", name, name)
		out.WriteString(stmt)
	}
	return out.String()

}

func NewNewTemplateVueCtx(vertical VerticalMeta) (NewTemplateVueCtx, error) {
	modelMeta := GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := NewTemplateVueCtx{
		ModelFieldsMeta:          modelMeta,
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		ModelTitleCaseName:       strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		CamelCasePluralModelName: strcase.ToLowerCamel(pName),
		TitleCaseModelPluralName: strcase.ToCamel(pName),
		FormMapStatment:          GenerateMapStatement(vertical.Model.Fields),          // TODO: LOOP {{.JSONFieldName}}:'{{.JSONFieldName}}',
		FormDefaultStatement:     GenerateFormDefaultsStatement(vertical.Model.Fields), // TODO:   {{.JSONFieldName}}:{{.JSONDefault}}, // default null|''|undefined|false
	}
	return out, nil
}

func GenerateNewVueFile(destDir, verticalName string, ctx NewTemplateVueCtx) (FileContainer, error) {
	modelName := strcase.ToLowerCamel(verticalName)
	newVueDest := path.Join(destDir, NewPaths().UIComponents)
	fileName := fmt.Sprintf("new%v.vue", modelName)

	b, err := ioutil.ReadFile("templates/ui/new-template.vue") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	newVuewTmpl := string(b)

	content, err := ExecuteTemplate("uiNewModel", newVuewTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: newVueDest,
		FileName:    fileName,
	}
	return out, nil
}

type ModelFieldMeta struct {
	FieldRule     string
	FieldName     string
	FieldLabel    string
	FieldType     string
	JSONFieldName string
	ColModifers   string
}
type ListTemplateVueCtx struct {
	ModelFieldsMeta          []ModelFieldMeta
	ModelTitleName           string
	TitleCaseModelName       string
	CamelCaseModelName       string
	ModelNamePluralTitleCase string
	CamelCasePlural          string
	COLOverrideStatement     string
	ResourceRoute            string
	FormDefaultStatement     string
	SearchStatement          string
}

func GetDefaultRule(t GoType) string {
	if t.IsNumeric() {
		return "min_value:1|numeric"
	}
	if t.IsString() {
		return "alpha_dash"
	}
	return "TODO: rule"
}

func GetFieldType(t GoType) string {
	if t.IsNumeric() {
		return "number"
	}
	if t.IsString() {
		return "text"
	}
	return "TODO: field type"
}

func GetColMod(t GoType) string {
	if t.IsNumeric() {
		return "\nnumberic\nsortable\n"
	}
	if t.IsString() {
		return "\nsortable\n"
	}
	return "TODO: field type"
}

func GetModelFieldMeta(vertical VerticalMeta) []ModelFieldMeta {
	out := []ModelFieldMeta{}
	for _, v := range vertical.Model.Fields {
		mfm := ModelFieldMeta{
			FieldRule:     GetDefaultRule(v),
			FieldName:     strcase.ToLowerCamel(v.Name),
			FieldLabel:    "TODO:" + v.Name,
			FieldType:     GetFieldType(v),
			ColModifers:   GetColMod(v),
			JSONFieldName: strcase.ToLowerCamel(v.Name),
		}
		out = append(out, mfm)
	}
	return out
}

func GenerateCOLOverrideStatement(fields []GoType) string {
	out := strings.Builder{}
	for _, v := range fields {
		leftCol := fmt.Sprintf("%s:'%s',\n", strcase.ToLowerCamel(v.Name), strcase.ToSnake(v.Name))
		rightCol := fmt.Sprintf("%s:'%s',\n", strcase.ToSnake(v.Name), strcase.ToLowerCamel(v.Name))
		out.WriteString(leftCol)
		out.WriteString(rightCol)
	}
	return out.String()
}

func GenerateFormDefaultsStatement(fields []GoType) string {
	out := strings.Builder{}
	for _, v := range fields {
		defstmt := fmt.Sprintf("%s: null,\n", strcase.ToLowerCamel(v.Name))
		out.WriteString(defstmt)
	}
	return out.String()
}

func GenerateSearchStatement(fields []GoType) string {
	out := strings.Builder{}

	for _, v := range fields {
		name := strcase.ToLowerCamel(v.Name)
		if v.IsNumeric() {
			stmt := fmt.Sprintf("\nif (this.search.%s && this.search.%s > 0) {\nfilter.$or.push({ %s: this.search.%s });\n}", name, name, name, name)
			out.WriteString(stmt)
		}
		if v.IsString() {
			txtTmpl := `
			if (this.search.{{.StringField}} && this.search.{{.StringField}}.length > 0) {
			let {{.StringField}} = this.search.{{.StringField}};
			filter.$or.push(
				{
					{{.StringField}}:{$like:` + "`%{{.StringField}}%`}\n}),\n}\n"

			stmt, err := ExecuteTemplate("searchTmpl", txtTmpl, map[string]interface{}{"StringField": name})
			if err != nil {
				println(err)
			}

			out.WriteString(stmt)
		}

	}
	return out.String()
}

func NewListTemplateVueCtx(vertical VerticalMeta) (ListTemplateVueCtx, error) {
	fieldsmeta := GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := ListTemplateVueCtx{
		ModelFieldsMeta:          fieldsmeta,
		COLOverrideStatement:     GenerateCOLOverrideStatement(vertical.Model.Fields),
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		FormDefaultStatement:     GenerateFormDefaultsStatement(vertical.Model.Fields),
		SearchStatement:          GenerateSearchStatement(vertical.Model.Fields),
		ModelTitleName:           strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		ModelNamePluralTitleCase: strcase.ToCamel(pName),
		CamelCasePlural:          strcase.ToLowerCamel(pName),
	}
	return out, nil
}

func GenerateListVueFile(destDir, verticalName string, ctx ListTemplateVueCtx) (FileContainer, error) {
	modelName := pluralizer.Plural(verticalName)
	modelName = strcase.ToLowerCamel(modelName)
	listVueDest := path.Join(destDir, NewPaths().UIComponents)
	fileName := fmt.Sprintf("%v.vue", modelName)

	b, err := ioutil.ReadFile("templates/ui/list-template.vue") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	listTmpl := string(b)

	content, err := ExecuteTemplate("uiListModel", listTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: listVueDest,
		FileName:    fileName,
	}
	return out, nil
}

type TestCTX struct {
	ImportPath               string
	ModelNameTitleCase       string
	ModelNamePluralTitleCase string
	ModelNamePluralCamel     string
	DefaultFieldStatement    string // TODO: {{.FieldGOName}}: {{.TODOStringOrINToRGODefault}},
}

func NewTestCtx(vertical VerticalMeta) (TestCTX, error) {
	pName := pluralizer.Plural(vertical.Name)
	out := TestCTX{
		ImportPath:               "",
		ModelNameTitleCase:       vertical.Name,
		ModelNamePluralTitleCase: pName,
		ModelNamePluralCamel:     strcase.ToCamel(pName),
		DefaultFieldStatement:    "//TODO: manually",
	}
	return out, nil
}

func GenerateRESTTestFile(destDir, verticalName string, ctx TestCTX) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	testfileDest := path.Join(destDir, NewPaths().Tests)
	fileName := fmt.Sprintf("%v_test.go", modelName)

	b, err := ioutil.ReadFile("templates/test-template.go") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	testTmpl := string(b)

	content, err := ExecuteTemplate("restCrudTest", testTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: testfileDest,
		FileName:    fileName,
	}
	return out, nil
}

type BaseAPPCTX struct {
	ImportPath string
}

func GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
	files, err := ReadFiles("static", destDir)
	if err != nil {
		return files, err
	}

	basics, err := ReadFiles("templates/basic", destDir)
	if err != nil {
		return files, err
	}

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
		{
			Name: "Records",
			Type: fmt.Sprintf("[]%v", structName),
			Tags: "json:\"records\"",
		},
		{
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
	CopyBase     bool
	Store        bool // dal/store/model.go
	Migrations   bool
	Controllers  bool
	Models       bool
	Utilities    bool
	JSStore      bool
	JSRouter     bool
	VueNewModel  bool
	VueListModel bool
	APICRUDTest  bool
}

func NewConfig() FeatureConfig {
	return FeatureConfig{
		CopyBase:     true,
		Store:        true,
		Migrations:   true,
		Controllers:  true,
		Models:       true,
		Utilities:    true,
		JSStore:      true,
		JSRouter:     true,
		VueNewModel:  true,
		VueListModel: true,
		APICRUDTest:  true,
	}
}

func GenerateTrim(structName string, stringFieldNames []string) (string, error) {
	const trimTmpl = `
	func (m *{{.StructName}}) Trim(){ {{ range  $value := .StringFieldNames }}
		m.{{$value}} = strings.TrimSpace(m.{{$value}}){{ end }}
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
	func (m {{.StructName}}) To{{.TargetName}}() {{.TargetName}} {
		out := {{.TargetName}}{{print "{}"}}
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
	func (m *{{.StructName}}) Initialize() { {{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
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

func GenerateNew(structName string, nilStatements map[string]string) (string, error) {
	const newTmpl = `
	func New{{.StructName}}()*{{.StructName}} {
		m := {{.StructName}}{}{{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
	return &m
	}`
	ctx := map[string]interface{}{
		"StructName":    structName,
		"NilStatements": nilStatements,
	}
	initFunc, err := ExecuteTemplate("newTmpl", newTmpl, ctx)
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

type Route struct {
	ResourceName       string
	PluralModelName    string
	TitleCaseModelName string
}

type JSRouterCTX struct {
	Routes []Route
}

func NewJSRouterCTX(verticals []VerticalMeta) (JSRouterCTX, error) {
	out := JSRouterCTX{
		Routes: []Route{},
	}

	for _, v := range verticals {
		pluralName := pluralizer.Plural(v.Name)
		route := Route{
			ResourceName:       strcase.ToKebab(v.Name),
			PluralModelName:    strcase.ToLowerCamel(pluralName),
			TitleCaseModelName: v.Name,
		}
		out.Routes = append(out.Routes, route)
	}
	return out, nil
}

func GenerateJSRouterFile(destDir string, ctx JSRouterCTX) (FileContainer, error) {
	routerDest := path.Join(destDir, NewPaths().UI)
	fileName := fmt.Sprintf("%v.js", "router") // TODO: magic strings

	b, err := ioutil.ReadFile("templates/ui/router-template.js") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	routerTmpl := string(b)

	content, err := ExecuteTemplate("uiRouter", routerTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: routerDest,
		FileName:    fileName,
	}
	return out, nil
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
