package glew

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

/*
* Vertical - is a vertical set of features around a single 'model'
 */

// Vertical Meta - all the meta information needed to create a vertical.
type VerticalMeta struct {
	Name        string
	Model       ModelMeta
	CreateModel ModelMeta
	UpdateModel ModelMeta
}

// GeneratedVertical - the resulting vertical feature set. Contains Raw strings and objects generated in case they are to be used elsewhere as well as file containers aka digital abstractions of files.
type GeneratedVertical struct {
	SQL              SQLStrings
	Controllers      []FileContainer
	Store            []FileContainer
	Models           []FileContainer
	Migrations       []FileContainer
	JSRouterTemplate FileContainer
	JSStores         []FileContainer
	VueNewModels     []FileContainer
	VueListModels    []FileContainer
	APICRUDTests     []FileContainer
}

// GenerateVerti calMeta - takes in model, name, and create/update reference models and then using reflection generates the meta for each struct.
// model must be populated. name can be empty, createM, putM can be nil.
func GenerateVerticalMeta(model interface{}, name string, createM, putM interface{}) (VerticalMeta, error) {
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

// GetSQLType - maps golang data type to corresponding postgres sql data type to be used in a create statement.
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

// NewDBTypeCtx - looks at the struct tags - `db`, and `db2` to get additional information or overrides of sql specific field flags.
// `db` Tag - specificy the column name. Hijaking db struct tag from pgx.
// `db2` Tag - custom tag from glew
// db2:type - specify or override the db column type.
// db2:default - specify the db default value if any.
// db2:pk - set to enable primary key statement
// db2:notnull - set to enable not null statement.

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

// NewSQLCtx - takes in the metadata for a given vertical and creates all related sql fields.
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

// Generates all the required
func NewStoreCtx(v VerticalMeta, sql SQLStrings, baseCtx BaseAPPCTX) StoreCtx {
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
		TODOProjectImportPath:    baseCtx.ImportPath,
		ModelNameTitleCase:       modelNameTitleCase,
		ModelNamePluralTitleCase: modelNamePluralTitleCase,
		TableName:                tableName,
		CreatePropertiesList:     createProperList,
		UpdatePropertiesList:     updateProperList,
		SQL:                      sql,
	}
	return out
}

func GetStringFields(fields []GoType) []string {
	out := []string{}
	for _, v := range fields {
		if v.IsString() {
			out = append(out, v.Name)
		}
	}
	return out
}

func GetNilableFields(fields []GoType) map[string]string {
	out := map[string]string{}
	for _, v := range fields {
		if v.IsNillable() {
			out[v.Name] = v.GetNewStatement()
		}
	}
	return out
}

//
func NewModelCtx(v VerticalMeta) (ModelCtx, error) {
	fields := []SField{}
	for _, v := range v.Model.Fields {
		fields = append(fields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	mName := v.Name
	model, err := GenerateStruct(mName, fields)
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
	updateName := fmt.Sprintf("Update%v", v.Name)
	updateModel, err := GenerateStruct(updateName, updateFields)
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
	createName := fmt.Sprintf("Create%v", v.Name)
	createModel, err := GenerateStruct(createName, createFields)
	if err != nil {
		return ModelCtx{}, err
	}
	utilities := ""
	page, err := GeneratePage(v.Name)
	if err != nil {
		return ModelCtx{}, err
	}
	mTrim, err := GenerateTrim(mName, GetStringFields(v.Model.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	updateTrim, err := GenerateTrim(updateName, GetStringFields(v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	createTrim, err := GenerateTrim(createName, GetStringFields(v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	mNil, err := GenerateInit(mName, GetNilableFields(v.Model.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	updateNil, err := GenerateInit(updateName, GetNilableFields(v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	createNil, err := GenerateInit(createName, GetNilableFields(v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	toUpdate, err := GenerateMapFunc(mName, updateName, GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	toCreate, err := GenerateMapFunc(mName, createName, GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	uToModel, err := GenerateMapFunc(updateName, mName, GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	uToCreate, err := GenerateMapFunc(updateName, createName, GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	cToModel, err := GenerateMapFunc(createName, mName, GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	cToUpdate, err := GenerateMapFunc(createName, updateName, GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	utilities = utilities + page + "\n"
	utilities = utilities + toUpdate + "\n"
	utilities = utilities + toCreate + "\n"
	utilities = utilities + uToModel + "\n"
	utilities = utilities + uToCreate + "\n"
	utilities = utilities + cToModel + "\n"
	utilities = utilities + cToUpdate + "\n"
	utilities = utilities + mTrim + "\n"
	utilities = utilities + mNil + "\n"
	utilities = utilities + updateTrim + "\n"
	utilities = utilities + updateNil + "\n"
	utilities = utilities + createTrim + "\n"
	utilities = utilities + createNil + "\n"
	out := ModelCtx{
		Model:       model,
		CreateModel: createModel,
		UpdateModel: updateModel,
		Utilities:   utilities,
	}
	return out, nil
}

// GenerateApp - takes in the required information to generate a basic crud app and based on the feature flags enabled, generates those features.
func GenerateApp(cfg FeatureConfig, destRoot, appName string, verticals []VerticalMeta, baseCtx BaseAPPCTX) ([]FileContainer, error) {
	// copy base
	destDir := destRoot //  filepath.Join(destRoot, "base-project")
	out := []FileContainer{}
	if cfg.CopyBase {
		files, err := GenerateBaseApp(destRoot, appName, baseCtx)
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
			ctx := NewStoreCtx(v, sql, baseCtx)
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
			ctx := NewControllerCtx(v.Name, baseCtx)
			controllerFile, err := GenerateControllerFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Controllers = append(verticalOut.Controllers, controllerFile)
		}
		if cfg.Models {
			ctx, err := NewModelCtx(v)
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

	if cfg.JSRouter {
		ctx, err := NewJSRouterCTX(verticals)
		if err != nil {
			return out, err
		}

	}

	for _, v := range verticalsOut {
		out = append(out, v.Migrations...)
		out = append(out, v.Controllers...)
		out = append(out, v.Store...)
		out = append(out, v.Models...)
	}

	return out, nil
}
