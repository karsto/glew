package glew

import (
	"reflect"

	"github.com/davecgh/go-spew/spew"
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
	updateProperties := []string{}
	for _, v := range v.UpdateModel.Fields {
		updateProperties = append(updateProperties, v.Name)
	}
	modelNamePluralTitleCase := pluralizer.Plural(modelNameTitleCase)

	out := StoreCtx{
		ModelNameTitleCase:       modelNameTitleCase,
		ModelNamePluralTitleCase: modelNamePluralTitleCase,
		TableName:                tableName,
		CreateProperties:         createProperties,
		UpdateProperties:         updateProperties,
		SQL:                      sql,
	}
	return out
}

func GenerateApp(destDir, appName string, verticals []VerticalMeta) ([]FileContainer, error) {
	// copy base
	cfg := NewConfig()
	out := []FileContainer{}
	if cfg.CopyBase {
		spew.Dump("FILES")
		files, err := GenerateBaseApp(destDir, appName)
		if err != nil {
			return out, err
		}
		spew.Dump(files)
		out = append(out, files...)
	}

	verticalsOut := []GeneratedVertical{}
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
			migrationStartId := 2 // TODO: read from directory automatically
			migrations, err := GenerateMigrationFiles(destDir, v.Name, sql, migrationStartId)
			if err != nil {
				return out, err
			}
			verticalOut.Migrations = append(verticalOut.Migrations, migrations...)
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
			ctx := ModelCtx{}
			model, err := GenerateModel(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Models = append(verticalOut.Models, model)
		}
		verticalsOut = append(verticalsOut, verticalOut)
	}

	for _, v := range verticalsOut {
		out = append(out, v.Controllers...)
		out = append(out, v.Store...)
		out = append(out, v.Models...)
	}

	return out, nil
}
