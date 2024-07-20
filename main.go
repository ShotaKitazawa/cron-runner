package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/adhocore/gronx"
	"github.com/spf13/cobra"
)

var config Config

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "cron-runner [flags] -- COMMAND [args...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate & construct
			argsLenAtDash := cmd.ArgsLenAtDash()
			if argsLenAtDash == -1 || argsLenAtDash == len(args) {
				return fmt.Errorf("COMMAND must be specified")
			}
			config.commands = args[argsLenAtDash:]
			if config.jobName != "" {
				return fmt.Errorf("job-name must not be empty")
			}
			if config.regex != "" {
				re, err := regexp.Compile(config.regex)
				if err != nil {
					return fmt.Errorf("regex is invalid: %w", err)
				}
				config.regexMatcher = re
			}
			if !gronx.New().IsValid(config.cronExpression) {
				return fmt.Errorf("cron expression is invalid")
			}
			// run
			return run()
		},
	}
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&config.cronExpression, "cron-expression", "c", "",
		"Cron schedule expression.")
	fs.BoolVarP(&config.ignoreExitCode, "ignore-exit-code", "", false,
		"Ignore exit code. Always succeed unless regex match fails.")
	fs.StringVarP(&config.jobName, "job-name", "n", "",
		"Job name for the label of exposed metrics")
	fs.StringVarP(&config.metricsBindAddr, "metrics-bind-addr", "", "0.0.0.0:9091",
		"bind address of HTTP server to expose metrics")
	fs.StringVarP(&config.output, "output", "o", "/dev/stdout",
		"Output file for the result of command")
	fs.StringVarP(&config.regex, "regex", "", "",
		"If regex does not matched to stdout and stderr, command is regarded as failure.")
	fs.Int64VarP(&config.timeoutMillisecond, "timeout-millisecond", "", 10000,
		"Timezone for cron expression. defaults to local.")
	fs.StringVarP(&config.timezone, "timezone", "", "",
		"Timezone for cron expression. defaults to local.")
	return rootCmd
}

func main() {
	command := NewCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
