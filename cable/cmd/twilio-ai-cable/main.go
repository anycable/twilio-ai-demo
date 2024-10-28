package main

import (
	"fmt"
	"os"

	acli "github.com/anycable/anycable-go/cli"
	"github.com/palkan/twilio-ai-cable/pkg/cli"
	"github.com/palkan/twilio-ai-cable/pkg/config"
	"github.com/palkan/twilio-ai-cable/pkg/version"

	_ "github.com/anycable/anycable-go/diagnostics"
)

func main() {
	conf := config.NewConfig()

	anyconf, err, ok := acli.NewConfigFromCLI(
		os.Args,
		acli.WithCLIName("mycable"),
		acli.WithCLIUsageHeader("MyCable, a custom AnyCable build"),
		acli.WithCLIVersion(version.Version()),
		acli.WithCLICustomOptions(cli.CustomOptions(conf)),
	)

	if err != nil {
		panic(err)
	}

	if ok {
		os.Exit(0)
	}

	if err := cli.Run(conf, anyconf); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
