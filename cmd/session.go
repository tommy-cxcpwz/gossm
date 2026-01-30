package cmd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tommy-cxcpwz/gossm/internal"
)

var (
	startSessionCommand = &cobra.Command{
		Use:   "start",
		Short: "Exec `start-session` under AWS SSM with interactive CLI",
		Long:  "Exec `start-session` under AWS SSM with interactive CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				target *internal.Target
				err    error
			)
			ctx := context.Background()
			ssmClient := ssm.NewFromConfig(*_credential.awsConfig)
			ec2Client := ec2.NewFromConfig(*_credential.awsConfig)

			// get target - if provided directly, skip the API lookup
			argTarget := strings.TrimSpace(viper.GetString("start-session-target"))
			if argTarget != "" {
				if err := internal.ValidateInstanceID(argTarget); err != nil {
					return err
				}
				target = &internal.Target{Name: argTarget}
			} else {
				target, err = internal.AskTarget(ctx, ssmClient, ec2Client)
				if err != nil {
					return err
				}
			}
			internal.PrintReady("start-session", _credential.awsConfig.Region, target.Name)

			input := &ssm.StartSessionInput{Target: aws.String(target.Name)}
			session, err := internal.CreateStartSession(ctx, ssmClient, input)
			if err != nil {
				return err
			}

			sessJson, err := json.Marshal(session)
			if err != nil {
				return err
			}

			paramsJson, err := json.Marshal(input)
			if err != nil {
				return err
			}

			if err := internal.CallProcess(_credential.ssmPluginPath, string(sessJson),
				_credential.awsConfig.Region, "StartSession",
				_credential.awsProfile, string(paramsJson)); err != nil {
				color.Red("%v", err)
			}

			if err := internal.DeleteStartSession(ctx, ssmClient, &ssm.TerminateSessionInput{
				SessionId: session.SessionId,
			}); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	startSessionCommand.Flags().StringP("target", "t", "", "[optional] it is ec2 instanceId.")
	viper.BindPFlag("start-session-target", startSessionCommand.Flags().Lookup("target"))

	// add sub command
	rootCmd.AddCommand(startSessionCommand)
}
