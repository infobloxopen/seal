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
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "seal",
	Short: "An opnionated authorization language, framework and runtime.",
	Long: `SEAL is a stack that allows developers to quickly develop
authorization for their applications. SEAL provides an
adaptable ecosystem that be targeted to a wide variety
applications.`,
	PersistentPreRunE: preRunE,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seal.yaml)")

	rootCmd.PersistentFlags().StringP("logging.level", "l", "info", "log level of application (debug, info, warn, error, fatal, panic)")
	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("logging.level"))

	rootCmd.PersistentFlags().StringP("logging.format", "", "text", "log format of application (text, json)")
	viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("logging.format"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".seal" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".seal")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// preRunE is executed before a subcommand is executed
func preRunE(cmd *cobra.Command, args []string) error {
	if err := setupLogger(); err != nil {
		return err
	}
	return nil
}

// setupLogger sets up the logger
func setupLogger() error {
	logger := logrus.StandardLogger()
	if viper.GetString("logging.format") == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: false})
	}

	// Set the log level on the default logger based on command line flag
	logLevels := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"warn":    logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}
	spec := viper.GetString("logging.level")
	if _, ok := logLevels[spec]; !ok {
		logger.WithField("logging.level", "info").Warnf("overrode invalid log level: %s", spec)
		spec = "info"
	} else {
		logger.WithField("logging.level", spec).Info("logging level")
	}
	logger.SetLevel(logLevels[spec])
	return nil
}
