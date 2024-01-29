package cmd

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var nextverCmd = &cobra.Command{
	Use:   "nextver",
	Short: "Prints the version of the next based on the commits since the last release.",
	RunE:  doNextver,
}

func init() {
	rootCmd.AddCommand(nextverCmd)
}

func doNextver(cmd *cobra.Command, args []string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	release, err := createRelease(r, "")
	if err != nil {
		return err
	}
	fmt.Printf("v%v\n", release.Version)
	return nil
}
