package glew

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

/*
* Vertical - is a vertical set of features around a single 'model'
 */

// Vertical Meta - all the meta information needed to create a vertical.
type VerticalMeta struct {
	Name        string //TODO: what is name vs model.name
	Model       ModelMeta
	CreateModel ModelMeta
	UpdateModel ModelMeta
}

// GeneratedVertical - the resulting vertical feature set. Contains Raw strings and objects generated in case they are to be used elsewhere as well as file containers aka digital abstractions of files.
type GeneratedVertical struct {
	SQL        SQLStrings
	Controller FileContainer
	Store      FileContainer
	Model      FileContainer
	Migrations []FileContainer

	JSStore      FileContainer
	VueNewModel  FileContainer
	VueListModel FileContainer
	APICRUDTests FileContainer
}

// GenerateVerticalMeta - takes in model, name, and create/update reference models and then using reflection generates the meta for each struct.
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
