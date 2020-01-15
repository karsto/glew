# glew

`glew` is a rapid prototyping and development tool, used to create starter web boiler-plate projects with CRUD+ functionality. The goal is to be able to create a set of models, feed it to glew, and have developer friendly flexible and ready to be developed against code come out. Its basically a glorified project template that makes some assumptions and 3rd party library choices to provide more functionality. This is not ready for external use.

The basic functionality out of the box:

* Complex multi app cli starter template.
  * Includes db creator + initializer wrapper.
  * Inlcudes Api only command + placeholders for `All` command.
  * Includes seprate run cmd and server for flexible ops.
* Basic error handling
  * panic handling.
  * hides internal errors.
  * parses certain db errors(FK,Unique,Check) as 400 status code.
* Gzip, recovery, console logger middlewares are enabled.
* Basic status controller endpoint check.
* Exposes expvar endpoint.
* Includes 'internal' utilities useful for developing apps.
* Includes drone build template.
* Includes docker-compose starter for local dev.
* Includes .gitignore and basic Makefile.
* All code supports a tenant id and enforces this very explicitly.

For each model provided a "feature" will be created. Each Feature has the following functionality:

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

Planned

* UI crud component based on vue / buefy with async paging and routes
* Complete base project that can run as output target
  * includes security / routing
* Generate test coverage
* Patch endpoint that supports batching.
* Put endpoint supports ETAG checks for pre put validation to verify user modified correct version.
* ETags - cache base middleware generator with class specific detail overrides.
* Cache - optional cache layer for db based on cache interface
* UI form gen
* UI validation gen
* .JS/.go client gen
* TenantId implementation changes to alternative? / make optional based on if present in model.
* auth,auth
* Support string (use 'code' convention)/ uuid for id field

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
| [vue.js](https://github.com/go-playground/validator)    |    base ui framework     |
| [bulma](https://bulma.io/)                              |    base css framework    |
| [buefy](https://buefy.org/)                             | buefy + bulma components |

### TODO:

- [ ] document with examples
- [ ] Add base tenant stuff
- [ ] would eventually like to make glew be able to target folder of models.
- [ ] enable config of meta tags used
- [ ] consider custom tags?
- [ ] .vue UI project
- [ ] Consider adding other recipes
  - [ ] web socket
  - [ ] rpc
  - [ ] services layer
  - [ ] ui admin