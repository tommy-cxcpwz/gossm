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
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			instanceID := args[0]
			command := strings.Join(args[1:], " ")
			skipCheck := viper.GetBool("exec-skip-check")

			// Validate instance ID format
			if !strings.HasPrefix(instanceID, "i-") {
				panicRed(fmt.Errorf("invalid instance ID format: %s (should start with 'i-')", instanceID))
			}

			// Check if instance is SSM-connected (unless skipped)
			if !skipCheck {
				connectedInstances, err := internal.FindInstanceIdsWithConnectedSSM(ctx, *_credential.awsConfig)
				if err != nil {
					panicRed(err)
				}

				found := false
				for _, id := range connectedInstances {
					if id == instanceID {
						found = true
						break
					}
				}
				if !found {
					panicRed(fmt.Errorf("instance %s is not connected to SSM.\nPossible causes:\n  - SSM agent is not running on the instance\n  - Instance lacks IAM permissions (AmazonSSMManagedInstanceCore)\n  - Network connectivity issues\n\nUse 'gossm list' to see available instances, or use --skip-check to bypass this validation", instanceID))
				}
			}

			target := &internal.Target{Name: instanceID}
			targets := []*internal.Target{target}

			internal.PrintReady(command, _credential.awsConfig.Region, instanceID)

			sendOutput, err := internal.SendCommand(ctx, *_credential.awsConfig, targets, command)
			if err != nil {
				panicRed(err)
			}

			fmt.Printf("%s\n", color.YellowString("Waiting for response..."))
			time.Sleep(time.Second * 2)

			// Get command result
			input := &ssm.GetCommandInvocationInput{
				CommandId:  sendOutput.Command.CommandId,
				InstanceId: aws.String(instanceID),
			}
			internal.PrintCommandInvocation(ctx, *_credential.awsConfig, []*ssm.GetCommandInvocationInput{input})
		},
	}
)

func init() {
	execCommand.Flags().Bool("skip-check", false, "[optional] skip SSM connectivity check before executing")
	viper.BindPFlag("exec-skip-check", execCommand.Flags().Lookup("skip-check"))

	rootCmd.AddCommand(execCommand)
}
