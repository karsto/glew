package main

import "time"

// test comment
// TODO: testing
type TestModel2 struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type CreateTestModel2 struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type UpdateTestModel2 struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
