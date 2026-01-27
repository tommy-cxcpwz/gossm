# gossm

`gossm` is an interactive CLI tool for connecting to AWS EC2 instances via AWS Systems Manager Session Manager.
<p align="center">
<img src="https://storage.googleapis.com/gjbae1212-asset/gossm/start.gif" width="500", height="450" />
</p>

<p align="center"/>
<a href="https://github.com/tommy-cxcpwz/gossm/actions/workflows/ci.yml"><img src="https://github.com/tommy-cxcpwz/gossm/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
<a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Ftommy-cxcpwz%2Fgossm"/></a>
<a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license" /></a>
<a href="https://goreportcard.com/report/github.com/tommy-cxcpwz/gossm"><img src="https://goreportcard.com/badge/github.com/tommy-cxcpwz/gossm" alt="Go Report Card"/></a>
</p>

## Overview

`gossm` is an interactive CLI tool for AWS Systems Manager Session Manager. It allows you to select EC2 instances with the SSM agent installed and connect to them using start-session, or execute commands remotely.

## Features

- **Interactive instance selection** - Browse and select from available EC2 instances
- **Start Session** - Connect to instances via SSM session manager
- **List Instances** - View all SSM-connected instances in a table format
- **Execute Commands** - Run commands on specific instances via SSM Run Command
- **Embedded SSM Plugin** - No need to install session-manager-plugin separately

## Prerequisite

### EC2 Instances
- [required] Your EC2 servers must have the [AWS SSM agent](https://docs.aws.amazon.com/systems-manager/latest/userguide/ssm-agent.html) installed
- [required] EC2 instances must have the **AmazonSSMManagedInstanceCore** IAM policy attached

### User Permissions
- [required] AWS access key and secret key
- [required] IAM permissions:
  - `ec2:DescribeInstances`
  - `ssm:StartSession`
  - `ssm:TerminateSession`
  - `ssm:DescribeSessions`
  - `ssm:DescribeInstanceInformation`
  - `ssm:DescribeInstanceProperties`
  - `ssm:GetConnectionStatus`
  - `ssm:SendCommand`
  - `ssm:GetCommandInvocation`
- [optional] `ec2:DescribeRegions` for region selection

## Install

### Homebrew
```bash
# install
$ brew tap tommy-cxcpwz/gossm
$ brew install gossm

# upgrade
$ brew upgrade gossm
```

### Download
[Download from releases](https://github.com/tommy-cxcpwz/gossm/releases)

## How to Use

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-p, --profile` | AWS profile name from credentials file | `default` or `AWS_PROFILE` env |
| `-r, --region` | AWS region | Interactive selection |
| `--debug` | Enable debug mode with timing information | `false` |

If no credentials file exists at `$HOME/.aws/credentials`, you can set `AWS_SHARED_CREDENTIALS_FILE` environment variable.

```ini
# credentials file format
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
```

### Commands

#### start

Start an interactive SSM session with a selected instance.

```bash
# Interactive mode - select instance from list
$ gossm start

# Direct mode - connect to specific instance
$ gossm start -t i-0abc123def456789
```

#### list

List all available instances that can be connected via SSM.

```bash
$ gossm list
```

Output shows instance name, ID, private DNS, and public DNS in a table format.

#### exec

Execute a command directly on a specific instance.

```bash
# Execute ls -la on a specific instance
$ gossm exec i-0abc123def456789 ls -la

# Execute a command with quotes
$ gossm exec i-0abc123def456789 "cat /etc/hosts"

# Check disk usage
$ gossm exec i-0abc123def456789 df -h

# Skip SSM connectivity check for faster execution
$ gossm exec --skip-check i-0abc123def456789 uptime
```

## LICENSE

This project is licensed under the MIT License.
