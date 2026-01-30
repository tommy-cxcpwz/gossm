package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tommy-cxcpwz/gossm/internal"
)

var (
	execCommand = &cobra.Command{
		Use:   "exec [command] [args...]",
		Short: "Execute a command on one or more instances via SSM",
		Long: `Execute a command on one or more instances via SSM.

Use -t/--target to specify instance IDs directly (repeatable), or omit to
interactively select multiple instances.

Examples:
  gossm exec --target i-0abc123def456789 ls -la
  gossm exec --target i-0abc123 --target i-0def456 "cat /etc/hosts"
  gossm exec df -h                  # interactive multi-select
  gossm exec --skip-check --target i-0abc123 ls -la`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ssmClient := ssm.NewFromConfig(*_credential.awsConfig)

			command := strings.Join(args, " ")
			skipCheck := viper.GetBool("exec-skip-check")
			targetFlags, _ := cmd.Flags().GetStringSlice("target")

			var targets []*internal.Target

			if len(targetFlags) > 0 {
				// Validate and build targets from flags
				for _, id := range targetFlags {
					if err := internal.ValidateInstanceID(id); err != nil {
						return err
					}
					targets = append(targets, &internal.Target{Name: id})
				}

				// Check SSM connectivity unless skipped
				if !skipCheck {
					connectedInstances, err := internal.FindInstanceIdsWithConnectedSSM(ctx, ssmClient)
					if err != nil {
						return err
					}
					connSet := make(map[string]struct{}, len(connectedInstances))
					for _, id := range connectedInstances {
						connSet[id] = struct{}{}
					}
					for _, t := range targets {
						if _, ok := connSet[t.Name]; !ok {
							return fmt.Errorf("instance %s is not connected to SSM.\nPossible causes:\n  - SSM agent is not running on the instance\n  - Instance lacks IAM permissions (AmazonSSMManagedInstanceCore)\n  - Network connectivity issues\n\nUse 'gossm list' to see available instances, or use --skip-check to bypass this validation", t.Name)
						}
					}
				}
			} else {
				// Interactive multi-select
				ec2Client := ec2.NewFromConfig(*_credential.awsConfig)
				selected, err := internal.AskMultiTarget(ctx, ssmClient, ec2Client)
				if err != nil {
					return err
				}
				targets = selected
			}

			internal.PrintReadyMulti(command, _credential.awsConfig.Region, targets)

			sendOutput, err := internal.SendCommand(ctx, ssmClient, targets, command)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", color.YellowString("Waiting for response..."))
			time.Sleep(time.Second * 2)

			// Build invocation inputs and name map for all targets
			inputs := make([]*ssm.GetCommandInvocationInput, 0, len(targets))
			nameMap := make(map[string]string, len(targets))
			for _, t := range targets {
				inputs = append(inputs, &ssm.GetCommandInvocationInput{
					CommandId:  sendOutput.Command.CommandId,
					InstanceId: aws.String(t.Name),
				})
				nameMap[t.Name] = t.TagName
			}
			internal.PrintCommandInvocation(ctx, ssmClient, inputs, nameMap)
			return nil
		},
	}
)

func init() {
	execCommand.Flags().StringSliceP("target", "t", nil, "target instance ID (repeatable)")
	execCommand.Flags().Bool("skip-check", false, "[optional] skip SSM connectivity check before executing")
	viper.BindPFlag("exec-skip-check", execCommand.Flags().Lookup("skip-check"))

	rootCmd.AddCommand(execCommand)
}
