package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tommy-cxcpwz/gossm/internal"
)

var (
	listCommand = &cobra.Command{
		Use:   "list",
		Short: "List all available instances that can be connected via SSM",
		Long:  "List all available instances that can be connected via SSM",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ssmClient := ssm.NewFromConfig(*_credential.awsConfig)
			ec2Client := ec2.NewFromConfig(*_credential.awsConfig)
			showTags, _ := cmd.Flags().GetBool("show-tags")

			table, err := internal.FindInstances(ctx, ssmClient, ec2Client)
			if err != nil {
				return err
			}

			if len(table) == 0 {
				color.Yellow("No instances found with SSM agent connected.")
				return nil
			}

			// Sort keys for consistent output
			keys := make([]string, 0, len(table))
			for k := range table {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			if showTags {
				// Print header
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, color.CyanString("NAME\tINSTANCE ID\tPRIVATE DNS\tPUBLIC DNS\tTAGS"))
				fmt.Fprintln(w, color.CyanString("----\t-----------\t-----------\t----------\t----"))
				w.Flush()

				for _, k := range keys {
					t := table[k]
					name, privateDNS, publicDNS := formatFields(t)

					fmt.Printf("%s  %s  %s  %s\n", name, t.Name, privateDNS, publicDNS)
					fmt.Printf("%s\n\n", internal.FormatTags(t.Tags))
				}
			} else {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, color.CyanString("NAME\tINSTANCE ID\tPRIVATE DNS\tPUBLIC DNS"))
				fmt.Fprintln(w, color.CyanString("----\t-----------\t-----------\t----------"))

				for _, k := range keys {
					t := table[k]
					name, privateDNS, publicDNS := formatFields(t)
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, t.Name, privateDNS, publicDNS)
				}
				w.Flush()
			}

			fmt.Printf("\n%s %d instance(s) found\n", color.GreenString("[OK]"), len(table))
			return nil
		},
	}
)

func formatFields(t *internal.Target) (name, privateDNS, publicDNS string) {
	name = t.TagName
	if name == "" {
		name = "-"
	}
	privateDNS = t.PrivateDomain
	if privateDNS == "" {
		privateDNS = "-"
	}
	publicDNS = t.PublicDomain
	if publicDNS == "" {
		publicDNS = "-"
	}
	return
}

func init() {
	listCommand.Flags().Bool("show-tags", false, "display instance tags")
	rootCmd.AddCommand(listCommand)
}
