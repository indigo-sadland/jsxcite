package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/indigo-sadland/jsxcite/internal/models"
	pathmatcher "github.com/indigo-sadland/jsxcite/internal/mods/path-matcher"
	"github.com/spf13/cobra"
)

var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "extract paths and URL-like strings from given JavaScript file(s)",
	Run: func(cmd *cobra.Command, args []string) {
		var paths []string

		target, _ := cmd.Flags().GetString("target")
		deepClean, _ := cmd.Flags().GetBool("deep-clean")

		// Handle '-t' flag first
		if target == "" {
			// Handle piped stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					paths = append(paths, scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					err := fmt.Errorf("reading stdin error: %v", err.Error())
					fmt.Println(err)
					os.Exit(1)
				}
			}
		} else {
			paths = append(paths, target)
		}

		// Check that at least one path is provided
		if len(paths) == 0 {
			err := fmt.Errorf("no input provided")
			fmt.Println(err)
			os.Exit(1)
		}

		var result []models.Path
		for _, path := range paths {
			ps, err := pathmatcher.ExtractPaths(path, deepClean)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			result = append(result, ps...)
		}

		if len(result) == 0 {
			err := fmt.Errorf("no paths found")
			fmt.Println(err)
			os.Exit(1)
		} else {
			for _, r := range result {
				res, _ := json.Marshal(r)
				fmt.Println(string(res))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pathsCmd)

	pathsCmd.Flags().StringP("target", "t", "", "URL or local path to target JS file\n"+
		"Examples: jsxcite paths -t https://example.com/libs/main.js\n "+
		"	jsxcite paths -t ./app/main.js\n")
	pathsCmd.Flags().BoolP("deep-clean", "c", true, "Filter out matched patterns "+
		"that most likely are false alarm. Enabled by default.")

}
