package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "builder",
    Short: "Builder tool for the C2 system",
    Long: `Builder tool for the C2 system.
This tool generates client executables based on the client module templates.`,
}

// Execute executes the root command
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
