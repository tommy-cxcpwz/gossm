package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tommy-cxcpwz/gossm/internal"
)

var (
	execCommand = &cobra.Command{
		Use:   "exec <instance-id> <command> [args...]",
		Short: "Execute a command on a specific instance via SSM",
		Long: `Execute a command on a specific instance via SSM.

Examples:
  gossm exec i-0abc123def456789 ls -la
  gossm exec i-0abc123def456789 "cat /etc/hosts"
  gossm exec i-0abc123def456789 df -h
  gossm exec --skip-check i-0abc123def456789 ls -la`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ssmClient := ssm.NewFromConfig(*_credential.awsConfig)

			instanceID := args[0]
			command := strings.Join(args[1:], " ")
			skipCheck := viper.GetBool("exec-skip-check")

			if err := internal.ValidateInstanceID(instanceID); err != nil {
				return err
			}

			// Check if instance is SSM-connected (unless skipped)
			if !skipCheck {
				connectedInstances, err := internal.FindInstanceIdsWithConnectedSSM(ctx, ssmClient)
				if err != nil {
					return err
				}

				found := false
				for _, id := range connectedInstances {
					if id == instanceID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("instance %s is not connected to SSM.\nPossible causes:\n  - SSM agent is not running on the instance\n  - Instance lacks IAM permissions (AmazonSSMManagedInstanceCore)\n  - Network connectivity issues\n\nUse 'gossm list' to see available instances, or use --skip-check to bypass this validation", instanceID)
				}
			}

			target := &internal.Target{Name: instanceID}
			targets := []*internal.Target{target}

			internal.PrintReady(command, _credential.awsConfig.Region, instanceID)

			sendOutput, err := internal.SendCommand(ctx, ssmClient, targets, command)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", color.YellowString("Waiting for response..."))
			time.Sleep(time.Second * 2)

			// Get command result
			input := &ssm.GetCommandInvocationInput{
				CommandId:  sendOutput.Command.CommandId,
				InstanceId: aws.String(instanceID),
			}
			internal.PrintCommandInvocation(ctx, ssmClient, []*ssm.GetCommandInvocationInput{input})
			return nil
		},
	}
)

func init() {
	execCommand.Flags().Bool("skip-check", false, "[optional] skip SSM connectivity check before executing")
	viper.BindPFlag("exec-skip-check", execCommand.Flags().Lookup("skip-check"))

	rootCmd.AddCommand(execCommand)
}
