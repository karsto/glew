package main

import (
	"fmt"
	"log"
	"os"

	cmd "{{.TODOProjectImportPath}}/cmd/cli"
)

// @title {{.AppName}}
// @version 1.0
// @description TODO
// @termsOfService TODO

// @contact.name API Support
// @contact.url TODO
// @contact.email support@todo.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host TODO
// @BasePath /v1
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
