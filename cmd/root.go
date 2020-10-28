package cmd

import (
	"github.com/karsto/glew/pkg"
	"github.com/spf13/cobra"
)

func init() {
	// rootCmd.AddCommand(versionCmd)

}

// TODO: add feature flags
// TODO: add optional directory
// TODO: add global config options
var RootCmd = &cobra.Command{
	Use:   "glew",
	Short: "glew",
	Long:  `glew`,

	Run: func(cmd *cobra.Command, args []string) {
		app := pkg.NewApp(pkg.Frontend{}, pkg.Backend{}, pkg.DB{})

		destDir := "../out"
		appName := "testApp"
		importPath := "github.com/karsto/glew/out"

		verticals, err := app.GetVerticalsFromFile("test-target.go")
		if err != nil {
			panic(err)
		}
		ctx := pkg.BaseAPPCTX{
			ImportPath: importPath, //TODO: hack to produce current import dir + out
		}

		features := pkg.NewConfig()
		files, err := app.GenerateApp(features, destDir, appName, verticals, ctx)

		if err != nil {
			panic(err)
		}
		err = pkg.WriteFiles(files, destDir)
		if err != nil {
			panic(err)
		}
	},
}
