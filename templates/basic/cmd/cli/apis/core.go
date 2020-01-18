package apis

import (
	"{{.TODOProjectImportPath}}/pkg/api/run/core"
	"github.com/spf13/cobra"
)

var coreAPICmd = &cobra.Command{
	Use:   "core",
	Short: "CRUD Web API for base-project, see help for env",
	Long: `CRUD Web API for base-project.
	ENV 			| Default
	PORT 			| 8080
	DBHOST 			| localhost
	PGPORT 			| 5432
	PGDATABASE 		| core
	PGUSER 			| postgres
	PGPASSWORD 		| postgres`,
	Run: func(cmd *cobra.Command, args []string) {
		core.Run()
	},
}

func init() {
	RootCmd.AddCommand(coreAPICmd)
}
