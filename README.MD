# glew
[![time tracker](https://wakatime.com/badge/github/karsto/glew.svg)](https://wakatime.com/badge/github/karsto/glew)

`glew` is a raw, rapid prototyping and development tool, used to create starter web boiler-plate projects with CRUD+ functionality. The goal is to be able to create a set of convention following models, feed it to glew, and have developer friendly flexible starter code ready produced. Its basically a glorified project template that makes some assumptions and 3rd party library choices to provide more functionality.

The basic functionality out of the box:


* Complete golang runtime with an API.
* Complex multi app cli pattern boiler plate.
  * Includes db creator + initializer wrapper.
  * Includes functionality for dynamic api stack.
  * Includes separate run cmd and server for flexible ops.
* API Basic error handling
  * panic handling.
  * hides internal errors.
  * parses certain db errors(FK,Unique,Check) as 400 status code.
* Gzip, recovery, console logger middlewares are enabled.
* Basic status controller endpoint check.
* Exposes expvar endpoint.
* Includes 'internal' utilities useful for developing apps.
* Includes `drone.yml` initial build template.
* Includes `docker-compose.yml` starter for local dev.
* Includes `.gitignore` and basic `Makefile`.
* All code supports a tenant id and enforces this very explicitly.
* All apps configured through env variables.

For each model set provided a "feature" will be created. Each Feature has the following functionality:

* A CRUD controller
  * All endpoints bind user input to data models.
  * All endpoints validate user input.
  * All endpoints have swagger documentation boiler plate text.
  * List endpoint supports paging and dynamic query filtering.
  * Delete endpoint supports multi delete.
* CRUD typed data store functions.
* Basic db migration up/down of table.
* Basic CRUD SQL statements.
* Utilities
  * Model initializer.
  * Model string trimmer.
  * Inter model mapping.
* All code is tenant aware.

Planned "Features":

* ETags aka web cache
* DAL Cache CompositePattern wrapping other Dal NewCache(*dal)..
* DAL interface + empty implementation generation
* CRUD `.go` tests
* CRUD `.go` api web client
* CRUD `.js` api web client
* CRUD `.MD` doc generation
* Model `.sql` insert dummy data statements
* Patch Batch singled field update endpoint for multi endpoints
  * patch send dictionary in map, allow multi param like delete
  * patch collection map[id][]patches with specific field updates
* PUT ETag version verification
* Support alt id type (string/uuid)
* UI CRUD Data store
* UI CRUD List Page
* UI CRUD Form Page
* UI CRUD Routing
* UI CRUD Testing

Planned Base features:

* UI CRUD
* Generic Cache interface
* Rate limiting
* Tenants CRUD API
* Optional/Alternative Tenant implementation
* Authentication api with multi bioler plate enabled
* auth,auth authboss, ladon, hydra
* user sign up portal + subscriptions
* search endpoints included in swagger documentation
* email templates
* notifications
* Metrics/Monitoring
* terrafrom templates for 'basic' aws

Glew TODO:

* better documentation
* cli - would like to target .go files

## Library Choices

glew relies heavily on 3rd party libraries for abstracting code and is mostly glue code between the libraries. Most of the applications data specific functionality is driven by the struct via tags.

These are the primary dependencies and what they are responsible for:

| Library                                                 |         Handles          |
| :------------------------------------------------------ | :----------------------: |
| [gin](https://github.com/gin-gonic/gin)                 |    base web framework    |
| [validator](https://github.com/go-playground/validator) |     input validation     |
| [sqlx](https://github.com/jmoiron/sqlx)                 |      sql interface       |
| [pgx](https://github.com/jackc/pgx)                     |     postgres driver      |
| [rql](https://github.com/a8m/rql)                       |   dynamic web queries    |
| [swag](https://github.com/swaggo/swag)                  |    swagger generator     |
| [migrate](https://github.com/golang-migrate/migrate)    |       db migration       |
| [viper](https://github.com/spf13/viper)                 |      configuration       |
| [cobra](https://github.com/spf13/cobra)                 |           cli            |
| [ladon](https://github.com/ory/ladon)                   |      authorization       |
| [authboss](https://github.com/volatiletech/authboss)    |      authentication      |
| [vue.js](https://github.com/go-playground/validator)    |    base ui framework     |
| [vue-router](https://github.com/vuejs/vue-router)       |    router                |
| [vuex](https://github.com/vuejs/vuex)                   |    state store           |
| [bulma](https://bulma.io/)                              |    base css framework    |
| [buefy](https://buefy.org/)                             | buefy + bulma components |

## CLI Structure

* `./` - all the dev tool ([drone](https://drone.io/), [go modules](https://blog.golang.org/using-go-modules), configs and main.go
* `./cmd` - the cli commands.
* `./cmd/cli/tools` - tools cli area.
* `./cmd/cli/apis` - for apis.
* `./db/migrations` - database migration scripts to be used with `go migrate` or `{{.CLINAME}} all` or `{{.CLINAME}} tools migrate`.
  * `/core` - schema for your models and related operations.
  * `/core_seed` - database seed insert script for quick testing.
* `./internal` - internal libraries and extension utilities for 3rd party libraries.
* `./pkg` - the applications relevant to the platform
  * `/apis` - all of the apis (core)

## Getting Started

Run the following application and then look at the output folder. You will find a new ready to run app, see `README.MD` for next steps.

```golang
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

func main() {
    // TODO: set your destination directory
	destDir := "out"
	verticals := []VerticalMeta{}
    appName:= "testApp"

	vertical, err := GenerateVertical(TestModel{}, "TestVertical", CreateTestModel{}, UpdateTestModel{})
	if err != nil {
		spew.Dump(err)
		panic(err)
	}

	verticals = append(verticals, vertical)
	ctx := BaseAPPCTX{
		ImportPath: "{{.TODOYourImportPath}}", // TODO: set your import path for the project to be generated ie github.com/karsto/test
	}
	files, err := GenerateApp(destDir, appName, verticals, ctx)
	if err != nil {
		spew.Dump(err)
		panic(err)
	}
	err = WriteFiles(files)
	if err != nil {
		spew.Dump(err)
		panic(err)
    }

    // TODO: ./out now has starter application with controllers and dal for given features. See README.md for next steps.
```