package cmd

import (
	"fmt"
	"os"

	"github.com/no8geo/notify/cmd/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Notify",
	Long:  `All software has versions. This is Notify's`,
	Run: func(cmd *cobra.Command, args []string) {
		yamlFile := "./chart/Chart.yaml"
		data, err := os.ReadFile(yamlFile)
		if err != nil {
			panic(err)
		}
		var c types.HelmChart
		err = yaml.Unmarshal(data, &c)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s Generator %v -- HEAD \n", c.Name, c.AppVersion)
	},
}
