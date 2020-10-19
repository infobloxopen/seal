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
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/infobloxopen/seal/pkg/compiler"
	// register the rego backend compiler
	_ "github.com/infobloxopen/seal/pkg/compiler/rego"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var compileSettings struct {
	files        []string // files to compile
	backend      string   // backend compiler
	outputFile   string   // output filename
	swaggerFiles []string // swagger file to read in types
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
	if len(compileSettings.swaggerFiles) == 0 {
		logrus.Fatal("swagger file is required for inferring types")
	}

	swaggerSpec := []string{}
	for _, file := range compileSettings.swaggerFiles {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			logrus.WithField("file", file).WithError(err).Fatal("could not read swagger file")
		}
		swaggerSpec = append(swaggerSpec, string(content))

	}

	cplr, err := compiler.NewPolicyCompiler(compileSettings.backend, swaggerSpec...)
	if err != nil {
		logrus.WithError(err).Fatal("could not create policy compiler")
	}

	var output []string
	for _, fil := range compileSettings.files {

		// read file
		input, err := ioutil.ReadFile(fil)
		if err != nil {
			logrus.WithField("file", fil).WithError(err).Fatal("could not read rules file")
		}

		// compile policies from policy rules
		pkgname := strings.TrimSuffix(path.Base(fil), ".seal")
		out, err := cplr.Compile(pkgname, string(input))
		if err != nil {
			logrus.WithField("file", fil).WithError(err).Fatal("could not compile rules file")
		}

		output = append(output, out)
	}

	// write to output
	switch compileSettings.outputFile {
	case "-", "":
		fmt.Printf("%s\n", strings.Join(output, "\n"))
	default:
		err := ioutil.WriteFile(compileSettings.outputFile,
			[]byte(fmt.Sprintf("%s\n", strings.Join(output, "\n"))),
			0644)
		if err != nil {
			logrus.WithField("file", compileSettings.outputFile).WithError(err).Fatal("could not write to output file")
		}
	}
}

func init() {
	rootCmd.AddCommand(compileCmd)

	// Here you will define your flags and configuration settings.
	compileCmd.PersistentFlags().StringArrayVarP(&compileSettings.files, "file", "f", []string{},
		"filename or directory to read seal files")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.backend, "backend", "b", "rego",
		"compiler backend")
	compileCmd.PersistentFlags().StringArrayVarP(&compileSettings.swaggerFiles, "swagger-file", "s", []string{},
		"filenames to read types")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.outputFile, "output", "o", "",
		"output file")
}
