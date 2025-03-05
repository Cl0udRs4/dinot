package cli

import (
    "fmt"

    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number of the Builder tool",
    Long:  `Print the version number of the Builder tool.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Builder tool v1.0.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
