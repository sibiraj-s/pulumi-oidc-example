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
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var pulumiDir = ".pulumi-backend"

var (
	region          = "ap-south-1"
	sessionName     = "PulumiLocalDev"
	awsProviderName = "awsProvider"
)

func getProvider(ctx *pulumi.Context) (*aws.Provider, error) {
	return aws.NewProvider(ctx, awsProviderName, &aws.ProviderArgs{
		// options can also be set as following
		Region: pulumi.String(region),
		AssumeRoleWithWebIdentity: &aws.ProviderAssumeRoleWithWebIdentityArgs{
			RoleArn:          pulumi.String(GetRoleArn()),
			WebIdentityToken: pulumi.String(GetOIDCToken()),
			SessionName:      pulumi.String(sessionName),
		},
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

	err := EnsureDir(CurrentDir(), pulumiDir)
	CheckErrX(err, fmt.Sprintf("Failed to create %s directory", pulumiDir))

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

	secretsProvider := auto.SecretsProvider("passphrase")
	envvars := auto.EnvVars(map[string]string{
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

	configOptions := auto.ConfigOptions{Path: true}
	configMap := auto.ConfigMap{
		"pulumi:disable-default-providers[0]": auto.ConfigValue{Value: "*"},
	}
	err = s.SetAllConfigWithOptions(ctx, configMap, &configOptions)
	CheckErrX(err, "Failed to disable default providers")

	w := s.Workspace()
	err = w.InstallPlugin(ctx, "aws", "v6.78.0")
	CheckErrX(err, "Failed to install aws plugin")

	_, err = s.Refresh(ctx, optrefresh.RunProgram(true))
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
