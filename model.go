package glew

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

type VerticalMeta struct {
	Name        string
	Model       ModelMeta
	CreateModel ModelMeta
	UpdateModel ModelMeta
}

type GeneratedVertical struct {
	SQL         SQLStrings
	Controllers []FileContainer
	Store       []FileContainer
	Models      []FileContainer
	Migrations  []FileContainer
}

// model must be populated. name can be empty, createM, putM can be nil.
func GenerateVertical(model interface{}, name string, createM, putM interface{}) (VerticalMeta, error) {
	if createM == nil {
		createM = model
	}
	if putM == nil && createM != nil {
		putM = createM
	}
	if putM == nil {
		putM = model
	}
	modelMeta, err := GetMeta(model)
	if err != nil {
		return VerticalMeta{}, err
	}
	if len(name) < 1 {
		name = modelMeta.Name
	}

	createMeta, err := GetMeta(createM)
	if err != nil {
		return VerticalMeta{}, err
	}
	updateMeta, err := GetMeta(putM)
	if err != nil {
		return VerticalMeta{}, err
	}
	out := VerticalMeta{
		Name:        name,
		Model:       modelMeta,
		CreateModel: createMeta,
		UpdateModel: updateMeta,
	}
	return out, nil
}

func GetSQLType(t reflect.Type) string {
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

func NewDBTypeCtx(t GoType) DBTypeCtx {
	name := t.Name
	if v, found := t.Tags.Lookup("db"); found {
		// TODO: nesting tags
		name = v
	}
	oType := GetSQLType(t.Type)
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

func NewSQLCtx(vertical VerticalMeta) SQLCtx {
	dbFields := []string{}
	createStatements := []string{}
	idColName := ".TODOidColName"
	for _, v := range vertical.Model.Fields {
		dbCtx := NewDBTypeCtx(v)
		if dbCtx.IsPK {
			idColName = dbCtx.Name
		}
		crtStmt := GenerateCreateStatement(dbCtx)
		createStatements = append(createStatements, crtStmt)
		dbFields = append(dbFields, dbCtx.Name)
	}

	insertFields := []string{}
	for _, v := range vertical.CreateModel.Fields {
		dbCtx := NewDBTypeCtx(v)
		insertFields = append(insertFields, dbCtx.Name)
	}

	putFields := []string{}
	for _, v := range vertical.CreateModel.Fields {
		dbCtx := NewDBTypeCtx(v)
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

func NewStoreCtx(v VerticalMeta, sql SQLStrings) StoreCtx {
	tableName := strcase.ToSnake(v.Name)
	modelNameTitleCase := strcase.ToCamel(v.Name)

	createProperties := []string{}
	for _, v := range v.CreateModel.Fields {
		createProperties = append(createProperties, v.Name)
	}
	listF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\tm.%v,\n", cur)
	}
	createProperList := AggStrList(createProperties, listF)
	createProperList = strings.Trim(createProperList, "\n")
	updateProperties := []string{}
	for _, v := range v.UpdateModel.Fields {
		updateProperties = append(updateProperties, v.Name)
	}
	updateProperList := AggStrList(updateProperties, listF)
	updateProperList = strings.Trim(updateProperList, "\n")
	modelNamePluralTitleCase := pluralizer.Plural(modelNameTitleCase)

	out := StoreCtx{
		ModelNameTitleCase:       modelNameTitleCase,
		ModelNamePluralTitleCase: modelNamePluralTitleCase,
		TableName:                tableName,
		CreatePropertiesList:     createProperList,
		UpdatePropertiesList:     updateProperList,
		SQL:                      sql,
	}
	return out
}

func NewModelCtx(v VerticalMeta, sql SQLStrings) (ModelCtx, error) {
	fields := []SField{}
	for _, v := range v.Model.Fields {
		fields = append(fields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	model, err := GenerateStruct(v.Name, fields)
	if err != nil {
		return ModelCtx{}, err
	}

	updateFields := []SField{}
	for _, v := range v.UpdateModel.Fields {
		updateFields = append(updateFields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	updateModel, err := GenerateStruct(fmt.Sprintf("Update%v", v.Name), updateFields)
	if err != nil {
		return ModelCtx{}, err
	}
	createFields := []SField{}
	for _, v := range v.CreateModel.Fields {
		createFields = append(createFields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	createModel, err := GenerateStruct(fmt.Sprintf("Create%v", v.Name), createFields)
	if err != nil {
		return ModelCtx{}, err
	}
	out := ModelCtx{
		Model:       model,
		CreateModel: createModel,
		UpdateModel: updateModel,
		Utilities:   "",
	}
	return out, nil
}

func GenerateApp(destRoot, appName string, verticals []VerticalMeta, ctx BaseAPPCTX) ([]FileContainer, error) {
	// copy base
	destDir := destRoot //  filepath.Join(destRoot, "base-project")
	cfg := NewConfig()
	out := []FileContainer{}
	if cfg.CopyBase {
		files, err := GenerateBaseApp(destRoot, appName, ctx)
		if err != nil {
			return out, err
		}
		out = append(out, files...)
	}

	verticalsOut := []GeneratedVertical{}
	migrationStartId := 2 // TODO: read from directory automatically

	for _, v := range verticals {
		verticalOut := GeneratedVertical{}

		ctx := NewSQLCtx(v)
		sql, err := GenerateSQL(ctx)
		if err != nil {
			return out, err
		}
		verticalOut.SQL = sql

		if cfg.Store {
			ctx := NewStoreCtx(v, sql)
			storeFile, err := GenerateStoreFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Store = append(verticalOut.Store, storeFile)
		}

		if cfg.Migrations {
			migrations, err := GenerateMigrationFiles(destDir, v.Name, sql, migrationStartId)
			if err != nil {
				return out, err
			}
			verticalOut.Migrations = append(verticalOut.Migrations, migrations...)
			migrationStartId += 1
		}

		if cfg.Controllers {
			ctx := NewControllerCtx(v.Name)
			controllerFile, err := GenerateControllerFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Controllers = append(verticalOut.Controllers, controllerFile)
		}
		if cfg.Models {
			ctx, err := NewModelCtx(v, sql)
			if err != nil {
				return out, err
			}
			model, err := GenerateModel(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Models = append(verticalOut.Models, model)
		}
		verticalsOut = append(verticalsOut, verticalOut)
	}

	for _, v := range verticalsOut {
		out = append(out, v.Migrations...)
		out = append(out, v.Controllers...)
		out = append(out, v.Store...)
		out = append(out, v.Models...)
	}

	return out, nil
}
