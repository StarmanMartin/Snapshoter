/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var (
	cfgFile, dst, src string
	r                 = regexp.MustCompile("Snapshot_")
)

func isDir(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}

	defer func(file *os.File) {
		cobra.CheckErr(file.Close())
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}

	// IsDir is short for fileInfo.Mode().IsDir()
	return fileInfo.IsDir()
}

func isDirValidate(path string) error {
	if isDir(path) {
		return nil
	}

	return errors.New("directory does not exist")
}

func manageArgs() {
	for _, name := range []string{"source", "destination"} {
		if !isDir(viper.GetString(name)) {
			prompt := promptui.Prompt{
				Label:    fmt.Sprintf("Enter %s directory", name),
				Validate: isDirValidate,
			}
			result, err := prompt.Run()
			cobra.CheckErr(err)
			viper.Set(name, result)
		}
	}

	if viper.GetInt("period") == 0 {
		prompt := promptui.Prompt{
			Label:    "Enter the time period to take a snapshot in hours directory",
			Validate: isDirValidate,
		}
		result, err := prompt.Run()
		cobra.CheckErr(err)
		viper.Set("period", result)
	}

	if viper.GetInt("max_shots") == 0 {
		prompt := promptui.Prompt{
			Label:    "How many Snapshots should be stored?",
			Validate: isDirValidate,
		}
		result, err := prompt.Run()
		cobra.CheckErr(err)
		viper.Set("max_shots", result)
	}

	err := viper.WriteConfig()
	cobra.CheckErr(err)
}

func cleanCurrentFolders() {
	files, err := os.ReadDir(dst)
	if err != nil {
		log.Fatal(err)
	}
	max_shots := viper.GetInt("max_shots")
	filenames := make([]string, 0)

	for _, file := range files {
		if r.MatchString(file.Name()) {
			filenames = append(filenames, file.Name())
		}
	}
	if len(filenames) > max_shots {
		sort.Strings(filenames)
		for _, filename := range filenames[0 : len(filenames)-max_shots+1] {
			torm := filepath.Join(dst, filename)
			err = os.RemoveAll(torm)
			cobra.CheckErr(err)
		}
	}

}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Snapshoter",
	Short: "Organizes snapshots of your filesystem",
	Long: `Snapshoter allows to take a snapshot of a source
directory at regular intervals. The cli arguments can be 
used to manage source path, target path and time period.`,

	Run: func(cmd *cobra.Command, args []string) {
		manageArgs()
		src = viper.GetString("source")
		dst = viper.GetString("destination")
		period := time.Duration(viper.GetInt("period"))
		dst = filepath.Join(dst, "Snapshoter")
		err := os.MkdirAll(dst, os.ModePerm)
		cobra.CheckErr(err)
		for true {
			current_time := time.Now()
			target := fmt.Sprintf("Snapshot_%s", current_time.Format("2006_01_02_15_04_05"))
			fmt.Println("New shot:", target)
			target = filepath.Join(dst, target)
			err = CopyDirectory(src, target)
			cobra.CheckErr(err)
			cleanCurrentFolders()
			time.Sleep(period * time.Hour)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to a YAML config file (default is $HOME/.cobra)")
	rootCmd.PersistentFlags().StringP("source", "s", "", "Source directory")
	rootCmd.PersistentFlags().StringP("destination", "d", "", "Source directory")
	rootCmd.PersistentFlags().IntP("period", "p", 0, "Snapshot period in hoers")
	rootCmd.PersistentFlags().IntP("max_shots", "m", 0, "The maximum number of snapshots to be stored")

	if err := viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source")); err != nil {
		return
	}

	if err := viper.BindPFlag("destination", rootCmd.PersistentFlags().Lookup("destination")); err != nil {
		return
	}

	if err := viper.BindPFlag("period", rootCmd.PersistentFlags().Lookup("period")); err != nil {
		return
	}

	if err := viper.BindPFlag("max_shots", rootCmd.PersistentFlags().Lookup("max_shots")); err != nil {
		return
	}

}

func initConfig() {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
