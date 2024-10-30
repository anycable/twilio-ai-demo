package cli

import (
	"github.com/palkan/twilio-ai-cable/pkg/config"
	"github.com/urfave/cli/v2"
)

func CustomOptions(conf *config.Config) func() ([]cli.Flag, error) {
	return func() ([]cli.Flag, error) {
		return []cli.Flag{
				&cli.StringFlag{
					Category:    "TWILIO",
					Name:        "twilio_account_sid",
					EnvVars:     []string{"TWILIO_ACCOUNT_SID"},
					Destination: &conf.Twilio.AccountSID,
				},
				&cli.BoolFlag{
					Category:    "MISC",
					Name:        "fake_rpc",
					EnvVars:     []string{"FAKE_RPC"},
					Destination: &conf.FakeRPC,
					Value:       conf.FakeRPC,
				},
			},
			nil
	}
}
