/*
Copyright © 2020 Infoblox <dev@infoblox.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"
	"log"

	"github.com/spf13/cobra"
	"github.com/infobloxopen/seal/pkg/compiler"
	_"github.com/infobloxopen/seal/pkg/compiler/rego"
)

var compileSettings struct {
	files []string // files to compile
	backend string // backend compiler
	outputFile string // output filename
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Takes a list of seal inputs and compiles them",
	Long: `compile takes a list of seal inputs
and compiles them to a target authorization runtime.
You can use one of the built in runtimes or call
a custom backend to target your own runtime.`,
	Run: compileFunc,
}

func compileFunc(cmd *cobra.Command, args []string) {

	_, err := compiler.New(compileSettings.backend)
	if err != nil {
		log.Printf("could not instantiate backend %v: %s", compileSettings.backend, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(compileCmd)

	// Here you will define your flags and configuration settings.
	compileCmd.PersistentFlags().StringArrayVarP(&compileSettings.files, "file", "f", []string{},
		"filename or diretory to get read seal files")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.backend, "backend", "b", "rego",
		"compiler backend")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.outputFile, "output", "o", "", 
		"output file")
}
