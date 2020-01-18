package glew

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	*/
	destDir := "out"
	verticals := []VerticalMeta{}

	vertical, err := GenerateVertical(TestModel{}, "TestVertical", CreateTestModel{}, UpdateTestModel{})
	if err != nil {
		spew.Dump(err)
		panic(err)
	}

	verticals = append(verticals, vertical)

	files, err := GenerateApp(destDir, "testApp", verticals)
	if err != nil {
		spew.Dump(err)
		panic(err)
	}
	err = WriteFiles(files)
	if err != nil {
		spew.Dump(err)
		panic(err)
	}
}
