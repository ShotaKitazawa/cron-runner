package main

import "regexp"

type Config struct {
	commands           []string
	cronExpression     string
	ignoreExitCode     bool
	jobName            string
	metricsBindAddr    string
	output             string
	regex              string
	regexMatcher       *regexp.Regexp
	timeoutMillisecond int64
	timezone           string
}
