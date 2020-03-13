package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const (
	exitCodeErrKnown   = 1
	exitCodeErrUnknown = 2
)

var cfgFile string
var verbose bool
var logger = logrus.WithField("service", "cluster_service")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cluster-service",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging (default is false)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func exitSuccess(message string) {
	fmt.Fprintf(os.Stdout, message)
	os.Exit(0)
}

func exitError(message string, code int) {
	fmt.Fprintf(os.Stderr, message)
	os.Exit(code)
}
