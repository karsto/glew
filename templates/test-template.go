// +build integration

package tests

import (
	"fmt"
	"net/http"
	"testing"

	"{{.ImportPath}}/pkg/api/model"
	"{{.ImportPath}}/pkg/api/tests/util"
)

type {{.ModelNameTitleCase}}Suite struct {
	resourcePath string
	name         string
}

func Test{{.ModelNamePluralTitleCase}}(t *testing.T) {
	{{.ModelNameTitleCase}}Suite := &{{.ModelNameTitleCase}}Suite{}
	{{.ModelNameTitleCase}}Suite.Run(t)
}

func (s *{{.ModelNameTitleCase}}Suite) Run(t *testing.T) {
	s.setup(t)
	t.Run(fmt.Sprintf("Can create %s", s.name), s.testCreate)
	t.Run(fmt.Sprintf("Can list %s", s.name), s.testList)
	t.Run(fmt.Sprintf("Can read %s", s.name), s.testRead)
	t.Run(fmt.Sprintf("Can update %s", s.name), s.testUpdate)
	t.Run(fmt.Sprintf("Can delete %s", s.name), s.testDelete)
	s.teardown(t)
}

func (s *{{.ModelNameTitleCase}}Suite) setup(t *testing.T) {
	s.name = "{{.ModelNamePluralCamel}}"
	s.resourcePath = fmt.Sprintf("%v/%v/", GlobalBasePath, s.name)
}

func (s *{{.ModelNameTitleCase}}Suite) teardown(t *testing.T) {}

func (s *{{.ModelNameTitleCase}}Suite) getNewModel() *model.{{.ModelNameTitleCase}} {
	return &model.{{.ModelNameTitleCase}}{
		{{.FieldGOName}}:{{.TODOStringOrINToRGODefault}}
	}
}

func (s *{{.ModelNameTitleCase}}Suite) testCreate(t *testing.T) {
	t.Parallel()
	s.createdAndValidate(t)
}

func (s *{{.ModelNameTitleCase}}Suite) testList(t *testing.T) {
	t.Parallel()
	created := s.createdAndValidate(t)
	res, err := GlobalClient.R().SetResult(model.{{.PluralTitleCaseModelName}}Page{}).Get(s.resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusOK {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusCreated, res.StatusCode(), res.Status())
	}
	page, success := res.Result().(*model.{{.PluralTitleCaseModelName}}Page)
	if !success {
		t.Fatal(fmt.Sprintf("unable to type assert %s page type for array", s.name))
	}

	found := false
	for _, f := range page.Records {
		if f.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created not found in list")
	}
}

func (s *{{.ModelNameTitleCase}}Suite) testRead(t *testing.T) {
	t.Parallel()
	s.createdAndValidate(t)
}

func (s *{{.ModelNameTitleCase}}Suite) testUpdate(t *testing.T) {
	t.Parallel()
	created := s.createdAndValidate(t)
	toUpdate := s.updateModel(t, created)

	res, err := GlobalClient.R().SetBody(toUpdate).SetResult(model.{{.ModelNameTitleCase}}{}).Put(fmt.Sprintf("%v%v", s.resourcePath, created.ID))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusOK {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusOK, res.StatusCode(), res.Status())
	}
	updatedModel, success := res.Result().(*model.{{.ModelNameTitleCase}})
	if !success {
		t.Fatalf("err result not expected type")
	}

	match, err := util.IsModelEqual(toUpdate, updatedModel, s.IgnoredFields(), "model not equal %v")
	if err != nil || !match {
		t.Fatal(err)
	}
}

func (s *{{.ModelNameTitleCase}}Suite) testDelete(t *testing.T) {
	t.Parallel()
	created := s.createdAndValidate(t)
	res, err := GlobalClient.R().Delete(fmt.Sprintf("%v%v", s.resourcePath, created.ID))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusNoContent {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusNoContent, res.StatusCode(), res.Status())
	}

	res, err = GlobalClient.R().Get(fmt.Sprintf("%v%v", s.resourcePath, created.ID))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusNotFound {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusNotFound, res.StatusCode(), res.Status())
	}
}

func (s *{{.ModelNameTitleCase}}Suite) IgnoredFields() []string {
	return []string{"createdAt", "updatedAt", "id"} // TODO: verity
}

func (s *{{.ModelNameTitleCase}}Suite) createdAndValidate(t *testing.T) model.{{.ModelNameTitleCase}} {
	toCreate := s.getNewModel()

	res, err := GlobalClient.R().SetBody(toCreate).SetResult(model.{{.ModelNameTitleCase}}{}).Post(s.resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusCreated {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusCreated, res.StatusCode(), res.Status())
	}
	created, success := res.Result().(*model.{{.ModelNameTitleCase}})
	if !success {
		t.Fatalf("err result not expected type")
	}
	// equality check the models to make sure they were correct
	match, err := util.IsModelEqual(toCreate, created, s.IgnoredFields(), "model not equal %v")
	if err != nil || !match {
		t.Fatal(err)
	}

	res, err = GlobalClient.R().SetResult(&model.{{.ModelNameTitleCase}}{}).Get(fmt.Sprintf("%v%v", s.resourcePath, created.ID))
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode() != http.StatusOK {
		t.Fatalf("expected %v status code found %v msg: %v ", http.StatusOK, res.StatusCode(), res.Status())
	}
	readResult, success := res.Result().(*model.{{.ModelNameTitleCase}})
	if !success {
		t.Fatalf("err result not expected type")
	}

	match, err = util.IsModelEqual(created, readResult, s.IgnoredFields(), "model not equal %v")
	if err != nil || !match {
		t.Fatal(err)
	}

	return *readResult
}

func (s *{{.ModelNameTitleCase}}Suite) updateModel(t *testing.T, m model.{{.ModelNameTitleCase}}) *model.{{.ModelNameTitleCase}} {
	// TODO:
	return &model.{{.ModelNameTitleCase}}{
		{{.FieldGOName}}:{{.TODOStringOrINToRGODefault}}
	}
}
