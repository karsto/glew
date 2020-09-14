package service

import (
"strings"

"github.com/karsto/glew/common/sqlutil"
"{{.TODOProjectImportPath}}/pkg/api/model"
)


type IService{{.ModelNamePluralTitleCase}} interface {
	Create(tenantID int, m model.Create{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	List(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{{.ModelNameTitleCase}}, int, error) {
	Read(tenantID, ID int) (model.{{.ModelNameTitleCase}}, error) {
	Update(tenantID, id int, m model.Update{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	Delete(tenantID int, IDs []int) (bool, error) {
}

func New{{.ModelNamePluralTitleCase}}Service({{.ModelNamePluralLowerCase}} ICRUD{{.ModelNameTitleCase}}) IService{{.ModelNamePluralTitleCase}} {
	out := {{.ModelNamePluralTitleCase}}Service{
		{{.ModelNamePluralLowerCase}}: {{.ModelNamePluralLowerCase}},
	}
	return out
}

type {{.ModelNamePluralTitleCase}}Service struct {
	{{.ModelNamePluralLowerCase}} ICRUD{{.ModelNameTitleCase}}
}

func (service *{{.ModelNamePluralTitleCase}}Service) Create(tenantID int, m model.Create{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	m, err := service.{{.ModelNamePluralLowerCase}}.Create(tenantID, m)
	return m, err
}

func (service *{{.ModelNamePluralTitleCase}}Service) List(tenantID, limit, offset int, sortExp, filterExp string, filterArgs []interface{}) ([]model.{{.ModelNameTitleCase}}, int, error) {
	result,total, err := service.{{.ModelNamePluralLowerCase}}.List(tenantID, limit, offset, sortExp, filterExp, filterArgs)
	return result, total, err
}

func (service *{{.ModelNamePluralTitleCase}}Service) Read(tenantID, ID int) (model.{{.ModelNameTitleCase}}, error) {
	m, err := service.{{.ModelNamePluralLowerCase}}.Read(tenantID, ID)
	return result, nil
}

func (service *{{.ModelNamePluralTitleCase}}Service) Update(tenantID, id int, m model.Update{{.ModelNameTitleCase}}) (model.{{.ModelNameTitleCase}}, error) {
	return service.{{.ModelNamePluralLowerCase}}.Update(tenantID, id, m)
}

func (service *{{.ModelNamePluralTitleCase}}Service) Delete(tenantID int, IDs []int) (bool, error) {
	didDelete, _, err := service.store.Delete(tenantID, IDS)
	return didDelete, err
}
