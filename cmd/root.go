/*
Copyright © 2023 Chris Collins 'collins.christopher@gmail.com'

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clcollins/srepd/pkg/pd"
	"github.com/clcollins/srepd/pkg/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cfgFile = "srepd.yaml"
const cfgFilePath = ".config/srepd/"

var PagerDutyOauthToken string
var Debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "srepd",
	Short: "TUI for common SREP PagerDuty on-call tasks",
	Long: `'srepd' is a TUI application for common PagerDuty 
on-call tasks.  It is intended to be used by SREs to perform 
such tasks as acknowledging incidents, adding notes, 
reassigning to the next on-call, etc.  It is not intended
to be a full-featured PagerDuty client, or kitchen sink, 
but rather a simple tool to make on-call tasks easier.`,

	Run: func(cmd *cobra.Command, args []string) {
		if Debug {
			for k, v := range viper.GetViper().AllSettings() {
				if k == "token" {
					v = "*****"
				}
				log.Printf("Found key: `%v`, value: `%v`\n", k, v)
			}
		}

		var ctx = context.Background()
		var tuiConfig = &tui.Config{
			Debug:   Debug,
			Verbose: false,
		}
		var pdConfig = &pd.Config{}

		token := viper.GetString("token")
		teams := viper.GetStringSlice("teams")
		silentuser := viper.GetString("silentuser")

		err := pdConfig.PopulateConfig(ctx, token, teams, silentuser)
		if err != nil {
			log.Fatal(err)
		}

		p := tea.NewProgram(tui.InitialModel(ctx, tuiConfig, pdConfig))
		_, err = p.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Enable debugging output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".srepd" (without extension).
	viper.AddConfigPath(home + "/" + cfgFilePath)
	viper.SetConfigName(cfgFile)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintln(os.Stderr, "Config file not found: "+err.Error())
		} else {
			fmt.Fprintln(os.Stderr, "Config file error: "+err.Error())
		}
	}
}
