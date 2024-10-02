package cmd

import (
  "os"
  _ "embed"

	"github.com/spf13/cobra"
)

//go:embed config/config.yaml
var config string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
    target := "config"

    if args.len == 1 {
      target = args[1]
    }

    if err := os.Mkdir(target, os.ModePerm); err != nil {
      log.Fatal(err)
    }

    if err := os.WriteFile(target, "/config.yaml", config, 0666); err != nil {
        log.Fatal(err)
    }
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
