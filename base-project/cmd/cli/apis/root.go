package apis

import (
	"github.com/karsto/base-project/pkg/api/run/all"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "apis",
	Short: "apis - runs all the apis (core).",
	Long:  `apis - runs all the apis (core).`,

	Run: func(cmd *cobra.Command, args []string) {
		all.Run()
	},
}
