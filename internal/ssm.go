package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssm_types "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/fatih/color"
)

const (
	maxOutputResults = 50
)

var (
	// default aws regions
	defaultAwsRegions = []string{
		"af-south-1",
		"ap-east-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3", "ap-south-1", "ap-southeast-2", "ap-southeast-3",
		"ca-central-1",
		"cn-north-1", "cn-northwest-1",
		"eu-central-1", "eu-north-1", "eu-south-1", "eu-west-1", "eu-west-2", "eu-west-3",
		"me-south-1",
		"sa-east-1",
		"us-east-1", "us-east-2", "us-gov-east-1", "us-gov-west-2", "us-west-1", "us-west-2",
	}
)

type (
	Target struct {
		Name          string
		PublicDomain  string
		PrivateDomain string
		displayKey    string // internal use for display formatting
	}

	Region struct {
		Name string
	}
)

// AskRegion asks you which selects a region.
func AskRegion(ctx context.Context, cfg aws.Config) (*Region, error) {
	var regions []string
	client := ec2.NewFromConfig(cfg)

	output, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		regions = make([]string, len(defaultAwsRegions))
		copy(regions, defaultAwsRegions)
	} else {
		regions = make([]string, len(output.Regions))
		for _, region := range output.Regions {
			regions = append(regions, aws.ToString(region.RegionName))
		}
	}
	sort.Strings(regions)

	var region string
	prompt := &survey.Select{
		Message: "Choose a region in AWS:",
		Options: regions,
	}
	if err := survey.AskOne(prompt, &region, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Format = "green+hb"
	}), survey.WithPageSize(20)); err != nil {
		return nil, err
	}

	return &Region{Name: region}, nil
}

// AskTarget asks you which selects an instance.
func AskTarget(ctx context.Context, cfg aws.Config) (*Target, error) {
	table, err := FindInstances(ctx, cfg)
	if err != nil {
		return nil, err
	}

	options := make([]string, 0, len(table))
	for k := range table {
		options = append(options, k)
	}
	sort.Strings(options)
	if len(options) == 0 {
		return nil, fmt.Errorf("not found ec2 instances")
	}

	prompt := &survey.Select{
		Message: "Choose a target in AWS:",
		Options: options,
	}

	selectKey := ""
	if err := survey.AskOne(prompt, &selectKey, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Format = "green+hb"
	}), survey.WithPageSize(20)); err != nil {
		return nil, err
	}

	return table[selectKey], nil
}

// FindInstances returns all of instances-map with running state.
func FindInstances(ctx context.Context, cfg aws.Config) (map[string]*Target, error) {
	timer := StartTimer("FindInstances")
	defer timer.Stop()

	var (
		ssmInstanceIDs []string
		ec2Instances   = make(map[string]*Target)
		ssmErr, ec2Err error
		wg             sync.WaitGroup
	)

	// Fetch SSM-connected instances in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		ssmTimer := StartTimer("SSM DescribeInstanceInformation")
		defer ssmTimer.Stop()

		client := ssm.NewFromConfig(cfg)
		input := &ssm.DescribeInstanceInformationInput{MaxResults: aws.Int32(maxOutputResults)}

		for {
			output, err := client.DescribeInstanceInformation(ctx, input)
			if err != nil {
				ssmErr = err
				return
			}
			for _, inst := range output.InstanceInformationList {
				ssmInstanceIDs = append(ssmInstanceIDs, aws.ToString(inst.InstanceId))
			}
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
	}()

	// Fetch EC2 instances in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		ec2Timer := StartTimer("EC2 DescribeInstances")
		defer ec2Timer.Stop()

		client := ec2.NewFromConfig(cfg)
		paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{
			Filters: []ec2_types.Filter{
				{Name: aws.String("instance-state-name"), Values: []string{"running"}},
			},
		})

		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				ec2Err = err
				return
			}
			for _, rv := range output.Reservations {
				for _, inst := range rv.Instances {
					instanceID := aws.ToString(inst.InstanceId)
					name := getInstanceName(inst.Tags)
					ec2Instances[instanceID] = &Target{
						Name:          instanceID,
						PublicDomain:  aws.ToString(inst.PublicDnsName),
						PrivateDomain: aws.ToString(inst.PrivateDnsName),
						displayKey:    fmt.Sprintf("%s\t(%s)", name, instanceID),
					}
				}
			}
		}
	}()

	wg.Wait()

	if ssmErr != nil {
		return nil, ssmErr
	}
	if ec2Err != nil {
		return nil, ec2Err
	}

	// Build result: only instances with SSM connected
	ssmSet := make(map[string]struct{}, len(ssmInstanceIDs))
	for _, id := range ssmInstanceIDs {
		ssmSet[id] = struct{}{}
	}

	result := make(map[string]*Target)
	for instanceID, target := range ec2Instances {
		if _, connected := ssmSet[instanceID]; connected {
			result[target.displayKey] = target
		}
	}

	return result, nil
}

// getInstanceName extracts the Name tag value from EC2 instance tags.
func getInstanceName(tags []ec2_types.Tag) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// FindInstanceIdsWithConnectedSSM asks you which selects instances.
func FindInstanceIdsWithConnectedSSM(ctx context.Context, cfg aws.Config) ([]string, error) {
	timer := StartTimer("SSM DescribeInstanceInformation")
	defer timer.Stop()

	var (
		instances  []string
		client     = ssm.NewFromConfig(cfg)
		outputFunc = func(instances []string, output *ssm.DescribeInstanceInformationOutput) []string {
			for _, inst := range output.InstanceInformationList {
				instances = append(instances, aws.ToString(inst.InstanceId))
			}
			return instances
		}
	)

	output, err := client.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{MaxResults: aws.Int32(maxOutputResults)})
	if err != nil {
		return nil, err
	}
	instances = outputFunc(instances, output)

	// Repeat it when if output.NextToken exists.
	if aws.ToString(output.NextToken) != "" {
		token := aws.ToString(output.NextToken)
		for {
			if token == "" {
				break
			}
			nextOutput, err := client.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{
				NextToken:  aws.String(token),
				MaxResults: aws.Int32(maxOutputResults)})
			if err != nil {
				return nil, err
			}
			instances = outputFunc(instances, nextOutput)

			token = aws.ToString(nextOutput.NextToken)
		}
	}

	return instances, nil
}

// CreateStartSession creates start session.
func CreateStartSession(ctx context.Context, cfg aws.Config, input *ssm.StartSessionInput) (*ssm.StartSessionOutput, error) {
	timer := StartTimer("SSM StartSession API")
	defer timer.Stop()

	client := ssm.NewFromConfig(cfg)
	return client.StartSession(ctx, input)
}

// DeleteStartSession creates session.
func DeleteStartSession(ctx context.Context, cfg aws.Config, input *ssm.TerminateSessionInput) error {
	client := ssm.NewFromConfig(cfg)
	fmt.Printf("%s %s \n", color.YellowString("Delete Session"),
		color.YellowString(aws.ToString(input.SessionId)))

	_, err := client.TerminateSession(ctx, input)
	return err
}

// SendCommand send commands to instance targets.
func SendCommand(ctx context.Context, cfg aws.Config, targets []*Target, command string) (*ssm.SendCommandOutput, error) {
	timer := StartTimer("SSM SendCommand API")
	defer timer.Stop()

	client := ssm.NewFromConfig(cfg)

	// only support to linux (window = "AWS-RunPowerShellScript")
	docName := "AWS-RunShellScript"

	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.Name)
	}

	input := &ssm.SendCommandInput{
		DocumentName:   &docName,
		InstanceIds:    ids,
		TimeoutSeconds: aws.Int32(60),
		CloudWatchOutputConfig: &ssm_types.CloudWatchOutputConfig{
			CloudWatchOutputEnabled: true,
		},
		Parameters: map[string][]string{"commands": {command}},
	}

	return client.SendCommand(ctx, input)
}

// PrintCommandInvocation watches command invocations.
func PrintCommandInvocation(ctx context.Context, cfg aws.Config, inputs []*ssm.GetCommandInvocationInput) {
	client := ssm.NewFromConfig(cfg)

	wg := new(sync.WaitGroup)
	for _, input := range inputs {
		wg.Add(1)
		go func(input *ssm.GetCommandInvocationInput) {
			defer wg.Done()
			for {
				// Check for context cancellation to prevent goroutine leak
				select {
				case <-ctx.Done():
					color.Yellow("[canceled] %s", aws.ToString(input.InstanceId))
					return
				default:
				}

				time.Sleep(1 * time.Second)
				output, err := client.GetCommandInvocation(ctx, input)
				if err != nil {
					color.Red("[err] %v", err)
					return
				}
				status := strings.ToLower(string(output.Status))
				switch status {
				case "pending", "inprogress", "delayed":
					continue
				case "success":
					stdout := aws.ToString(output.StandardOutputContent)
					if stdout == "" {
						stdout = "(no output)"
					}
					fmt.Printf("[%s][%s]\n%s\n", color.GreenString("success"), color.YellowString(aws.ToString(output.InstanceId)), stdout)
					return
				default:
					// Show both stdout and stderr for failed commands
					stdout := aws.ToString(output.StandardOutputContent)
					stderr := aws.ToString(output.StandardErrorContent)
					statusDetail := aws.ToString(output.StatusDetails)

					fmt.Printf("[%s][%s] status: %s\n", color.RedString("failed"), color.YellowString(aws.ToString(output.InstanceId)), color.RedString(statusDetail))
					if stdout != "" {
						fmt.Printf("stdout: %s\n", stdout)
					}
					if stderr != "" {
						fmt.Printf("stderr: %s\n", color.RedString(stderr))
					}
					if stdout == "" && stderr == "" {
						fmt.Printf("(no output)\n")
					}
					return
				}
			}
		}(input)
	}

	wg.Wait()
}

func PrintReady(cmd, region, target string) {
	fmt.Printf("[%s] region: %s, target: %s\n", color.GreenString(cmd), color.YellowString(region), color.YellowString(target))
}

// CallProcess calls process.
func CallProcess(process string, args ...string) error {
	call := exec.Command(process, args...)
	call.Stderr = os.Stderr
	call.Stdout = os.Stdout
	call.Stdin = os.Stdin

	// ignore signal(sigint)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	defer signal.Stop(sigs) // Properly unregister signal handler
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-sigs:
			case <-done:
				return // Use return instead of break to exit goroutine
			}
		}
	}()
	defer close(done)

	// run subprocess
	if err := call.Run(); err != nil {
		return WrapError(err)
	}
	return nil
}
