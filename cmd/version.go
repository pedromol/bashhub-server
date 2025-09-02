package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var (
	GitCommit  string
	BuildDate  string
	Version    string
	OsArch     = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	GoVersion  = runtime.Version()
	versionCmd = &cobra.Command{

		Use:   "version",
		Short: "Print the version number and build info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Build Date:", BuildDate)
			fmt.Println("Git Commit:", GitCommit)
			fmt.Println("Version:", Version)
			fmt.Println("Go Version:", GoVersion)
			fmt.Println("OS / Arch:", OsArch)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
