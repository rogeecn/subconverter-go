package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/subconverter/subconverter-go/internal/app/converter"
	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
)

var (
	cfgFile   string
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "subctl",
	Short: "SubConverter CLI - Convert proxy subscriptions from command line",
	Long: `SubConverter CLI is a command-line tool for converting proxy subscription URLs
between different formats like Clash, Surge, Quantumult, etc.`,
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert subscription URLs to target format",
	Long:  `Convert subscription URLs to the specified target format`,
	Args:  cobra.MinimumNArgs(1),
	Run:   runConvert,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate subscription URLs",
	Long:  `Validate subscription URLs and check their format`,
	Args:  cobra.MinimumNArgs(1),
	Run:   runValidate,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show supported formats and features",
	Long:  `Display information about supported formats and features`,
	Run:   runInfo,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./configs/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format (text, json)")

	convertCmd.Flags().StringP("target", "t", "clash", "target format (clash, surge, quantumult, loon, v2ray, surfboard)")
	convertCmd.Flags().StringP("config", "c", "", "custom configuration file")
	convertCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	convertCmd.Flags().StringSliceP("include", "i", []string{}, "include proxies with remarks matching patterns")
	convertCmd.Flags().StringSliceP("exclude", "e", []string{}, "exclude proxies with remarks matching patterns")
	convertCmd.Flags().Bool("udp", true, "enable UDP support")
	convertCmd.Flags().Bool("sort", true, "sort proxies by name")
	convertCmd.Flags().Bool("insecure", false, "skip TLS verification")

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(infoCmd)
}

func initConfig() {
	config.Load()
}

func runConvert(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	
	log := logger.New(logger.Config{
		Level:  logLevel,
		Format: logFormat,
		Output: "stdout",
	})

	target, _ := cmd.Flags().GetString("target")
	configURL, _ := cmd.Flags().GetString("config")
	output, _ := cmd.Flags().GetString("output")
	include, _ := cmd.Flags().GetStringSlice("include")
	exclude, _ := cmd.Flags().GetStringSlice("exclude")
	udp, _ := cmd.Flags().GetBool("udp")
	sort, _ := cmd.Flags().GetBool("sort")

	service := converter.NewService(cfg, log)
	service.RegisterGenerators()

	req := &converter.ConvertRequest{
		Target:    target,
		URLs:      args,
		ConfigURL: configURL,
		Options: converter.Options{
			IncludeRemarks: include,
			ExcludeRemarks: exclude,
			UDP:           udp,
			Sort:          sort,
		},
	}

	resp, err := service.Convert(cmd.Context(), req)
	if err != nil {
		log.WithError(err).Fatal("Failed to convert subscription")
	}

	if output != "" {
		if err := os.WriteFile(output, []byte(resp.Config), 0644); err != nil {
			log.WithError(err).Fatal("Failed to write output file")
		}
		log.Infof("Configuration saved to %s", output)
	} else {
		fmt.Println(resp.Config)
	}
}

func runValidate(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	log := logger.New(logger.Config{
		Level:  logLevel,
		Format: logFormat,
		Output: "stdout",
	})

	service := converter.NewService(cfg, log)
	service.RegisterGenerators()

	for _, url := range args {
		req := &converter.ValidateRequest{
			URL: url,
		}

		resp, err := service.Validate(cmd.Context(), req)
		if err != nil {
			log.WithError(err).Errorf("Failed to validate URL: %s", url)
			continue
		}

		fmt.Printf("URL: %s\n", url)
		fmt.Printf("Valid: %t\n", resp.Valid)
		if resp.Error != "" {
			fmt.Printf("Error: %s\n", resp.Error)
		}
		if resp.Valid {
			fmt.Printf("Format: %s\n", resp.Format)
			fmt.Printf("Proxies: %d\n", resp.Proxies)
		}
		fmt.Println()
	}
}

func runInfo(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	log := logger.New(logger.Config{
		Level:  logLevel,
		Format: logFormat,
		Output: "stdout",
	})

	service := converter.NewService(cfg, log)
	service.RegisterGenerators()

	resp, err := service.GetInfo(cmd.Context())
	if err != nil {
		log.WithError(err).Fatal("Failed to get service info")
	}

	fmt.Printf("SubConverter CLI v%s\n", resp.Version)
	fmt.Println("Supported formats:")
	for _, format := range resp.SupportedTypes {
		fmt.Printf("  - %s\n", format)
	}
	fmt.Println("Features:")
	for _, feature := range resp.Features {
		fmt.Printf("  - %s\n", feature)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}