package glew

import (
	"testing"
	"time"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

type TestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type CreateTestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type UpdateTestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}

func TestPlayground(t *testing.T) {
	/*
		CONFIG
		destDir - the output directory from this run
		appName - the name of the application and docker runtime
		importPath - application go import directory, aka the directory of the app to reference itself
	*/
	app := App{
		db:       DB{},
		frontend: Frontend{},
		backend:  Backend{},
	}
	destDir := "out"
	appName := "testApp"
	importPath := "github.com/ashtonian/glew/out"

	verticals := []VerticalMeta{}

	vertical, err := app.GenerateVerticalMeta(TestModel{}, "TestVertical", CreateTestModel{}, UpdateTestModel{})
	if err != nil {
		panic(err)
	}

	verticals = append(verticals, vertical)

	ctx := BaseAPPCTX{
		ImportPath: importPath, //TODO: hack to produce current import dir + out
	}

	features := NewConfig()
	files, err := app.GenerateApp(features, destDir, appName, verticals, ctx)

	if err != nil {
		panic(err)
	}
	err = WriteFiles(files)
	if err != nil {
		panic(err)
	}
}
