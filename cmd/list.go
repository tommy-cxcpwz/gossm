package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tommy-cxcpwz/gossm/internal"
)

var (
	listCommand = &cobra.Command{
		Use:   "list",
		Short: "List all available instances that can be connected via SSM",
		Long:  "List all available instances that can be connected via SSM",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			table, err := internal.FindInstances(ctx, *_credential.awsConfig)
			if err != nil {
				panicRed(err)
			}

			if len(table) == 0 {
				color.Yellow("No instances found with SSM agent connected.")
				return
			}

			// Sort keys for consistent output
			keys := make([]string, 0, len(table))
			for k := range table {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// Print header
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, color.CyanString("NAME\tINSTANCE ID\tPRIVATE DNS\tPUBLIC DNS"))
			fmt.Fprintln(w, color.CyanString("----\t-----------\t-----------\t----------"))

			// Print instances
			for _, k := range keys {
				t := table[k]
				// Extract name from key (format: "name\t(instance-id)")
				name := ""
				if idx := len(k) - len(t.Name) - 3; idx > 0 {
					name = k[:idx]
				}

				publicDNS := t.PublicDomain
				if publicDNS == "" {
					publicDNS = "-"
				}
				privateDNS := t.PrivateDomain
				if privateDNS == "" {
					privateDNS = "-"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, t.Name, privateDNS, publicDNS)
			}
			w.Flush()

			fmt.Printf("\n%s %d instance(s) found\n", color.GreenString("[OK]"), len(table))
		},
	}
)

func init() {
	rootCmd.AddCommand(listCommand)
}
