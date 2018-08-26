// Copyright © 2018 K. Alex Mills <k.alex.mills@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/kalexmills/collabbook-go/data"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
	"github.com/kalexmills/collabbook-go/view"
)

var cfgFile string

// itemstore is a global repo that's typically loaded by the root command and made available to other commands.
var itemstore *data.Repo

var cbPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cb",
	Short: "Tasks, boards & notes for the command-line habitat",
	DisableFlagsInUseLine: true,
	Long: `
   TODO: Long description.
`,
	// PersistentPreRun crawls up the working directory, checking for a .collabbook file and loading it when it finds it.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()

		for !os.IsNotExist(err) && !os.IsPermission(err) {
			cbPath = filepath.Join(wd, ".collabbook")
			var info os.FileInfo
			info, err = os.Lstat(cbPath)

			if err == nil && info.Mode().IsRegular() {
				var bytes []byte
				bytes, err = ioutil.ReadFile(cbPath)

				itemstore = data.NewRepo()
				err = itemstore.UnmarshalText(bytes)
				if err != nil {
					fmt.Printf("Corrupted .collabbook file found at " + filepath.Join(wd, ".collabbook"))
					os.Exit(1)
				}
				return
			}

			wd, _ = filepath.Split(wd)
			wd = filepath.Clean(wd)
			info, err = os.Lstat(wd)

			if wd == filepath.VolumeName(wd)+string(filepath.Separator) {
				fmt.Printf("Could not find .collabbook file in any ancestor directory. Stopping at filesystem boundary.")
				os.Exit(1)
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if bytes, err := itemstore.MarshalText(); err == nil {
			ioutil.WriteFile(cbPath, bytes, 0644)
		} else {
			view.Failure(":-O", "Could not write file because:\n\t" + err.Error())
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".collabbook.conf")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
