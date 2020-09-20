package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// rootCmd.AddCommand(versionCmd)

}

var RootCmd = &cobra.Command{
	Use:   "glew",
	Short: "glew",
	Long:  `glew`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("t")
	},
}
