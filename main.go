package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var pulumiDir = ".pulumi"

var (
	region      = "ap-south-1"
	sessionName = "PulumiLocalDev"
)

func getProvider(ctx *pulumi.Context) (*aws.Provider, error) {
	return aws.NewProvider(ctx, "awsProvider", &aws.ProviderArgs{
		// // options can also be set as following
		// Region: pulumi.String(region),
		// AssumeRoleWithWebIdentity: &aws.ProviderAssumeRoleWithWebIdentityArgs{
		// 	RoleArn:              pulumi.String(GetRoleArn()),
		// 	WebIdentityTokenFile: pulumi.String(TokenFilePath()),
		// 	SessionName:          pulumi.String(sessionName),
		// },
	})
}

func deployFunc(ctx *pulumi.Context) error {
	provider, err := getProvider(ctx)
	if err != nil {
		return err
	}

	bucketArgs := &s3.BucketV2Args{}
	bucket, err := s3.NewBucketV2(ctx, "TestBucket", bucketArgs, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	ctx.Export("bucketName", bucket.ID())
	return nil
}

func main() {
	// to destroy our program, we can run `go run main.go destroy`
	destroy := false
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "destroy" {
			destroy = true
		}
	}

	ctx := context.Background()
	currentDir := CurrentDir()

	projectName := "dev"
	stackName := "createS3Bucket"

	// fullStackName := auto.FullyQualifiedStackName("local", projectName, stackName)
	// fmt.Println("Stack name:", fullStackName)

	err := EnsureDir(CurrentDir(), pulumiDir)
	CheckErrX(err, "Failed to create .pulumi directory")

	// get the oidc token,
	// this will write the token to a file
	GetOIDCToken()

	// set the Pulumi home directory to the current directory for easier debugging during development.
	pulumiHomeDir := filepath.Join(currentDir, pulumiDir)
	pulumiBackendURL := "file://" + pulumiHomeDir

	os.Setenv("PULUMI_HOME", pulumiHomeDir)

	// specify a local backend for Pulumi instead of using the Pulumi service.
	project := auto.Project(workspace.Project{
		Name:        tokens.PackageName(projectName),
		Description: StringPtr("This is a test project to create an s3 bucket"),
		Runtime:     workspace.NewProjectRuntimeInfo("go", nil),
		Backend: &workspace.ProjectBackend{
			URL: pulumiBackendURL,
		},
	})

	// setup a passphrase secrets provider and configure environment variables.
	secretsProvider := auto.SecretsProvider("passphrase")
	envvars := auto.EnvVars(map[string]string{
		// in a real program, securely provide the password or use the actual environment.
		"PULUMI_CONFIG_PASSPHRASE": "password",
		"PULUMI_HOME":              os.Getenv("PULUMI_HOME"),
	})
	stackSettings := auto.Stacks(map[string]workspace.ProjectStack{
		stackName: {SecretsProvider: "passphrase"},
	})

	workspaceOpts := []auto.LocalWorkspaceOption{
		project,
		secretsProvider,
		envvars,
		stackSettings,
	}
	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc, workspaceOpts...)
	CheckErrX(err, "Failed to setup workspace")

	// set configuration values with path options if required.
	configOptions := auto.ConfigOptions{Path: true}
	configMap := auto.ConfigMap{
		"pulumi:disable-default-providers[0]": auto.ConfigValue{Value: "*"}, // disable all default providers; the AWS provider will be installed manually.
		// // avoid setting these values in config file,
		// // refer to the README for more details.
		// "aws:region": auto.ConfigValue{Value: region},
		// "aws:assumeRoleWithWebIdentity.roleArn":          auto.ConfigValue{Value: GetRoleArn(), Secret: true},
		// "aws:assumeRoleWithWebIdentity.webIdentityToken": auto.ConfigValue{Value: GetOIDCToken(), Secret: true},
		// "aws:assumeRoleWithWebIdentity.sessionName":      auto.ConfigValue{Value: sessionName},
	}
	err = s.SetAllConfigWithOptions(ctx, configMap, &configOptions)
	CheckErrX(err, "Failed to disable default providers")

	// install the AWS provider plugin.
	w := s.Workspace()
	err = w.InstallPlugin(ctx, "aws", "v6.75.0")
	CheckErrX(err, "Failed to install aws plugin")

	// just demonstration for how env vars can be set on the workspace
	// we directly pass them to the provider in the getProvider function
	awsEnvVars := map[string]string{
		"AWS_REGION":                  region,
		"AWS_ROLE_ARN":                GetRoleArn(),
		"AWS_WEB_IDENTITY_TOKEN_FILE": TokenFilePath(),
		"AWS_ROLE_SESSION_NAME":       sessionName,
	}
	err = w.SetEnvVars(awsEnvVars)
	CheckErrX(err, "Failed to set env vars on workspace")

	// refresh the stack to ensure the state is up-to-date.
	_, err = s.Refresh(ctx)
	CheckErrX(err, "Failed to refresh stack")

	if destroy {
		fmt.Println("Starting stack destroy")

		// wire up our destroy to stream progress to stdout
		stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)

		// destroy our stack and exit early
		_, err := s.Destroy(ctx, stdoutStreamer, optdestroy.Remove())
		CheckErrX(err, "Failed to destroy stack")

		fmt.Println("Stack successfully destroyed")
		os.Exit(0)
	}

	// wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	// execute the update operation on the stack.
	_, err = s.Up(ctx, stdoutStreamer)
	CheckErrX(err, "Failed to update stack")
}
