# Pulumi OIDC Example

This is an example of how to use OIDC with Pulumi automation api to assume a role in AWS and create resources.

## Prerequisites

- Go (check the go version in the `go.mod` file)

## Create resources

Export `ROLE_ARN` and `OIDC_TOKEN` as environment variables and run the program.

> [!NOTE]
> In real world, the token would be fetched from the OIDC provider dynamically. For testing, we are just using a static token.

```bash
export ROLE_ARN="arn:aws:iam::123456789012:role/TestRole"
export OIDC_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...."
go run .
```

This will create an S3 bucket in the AWS account by assuming the role using the provided OIDC token.

Also, the Pulumi home directory and the project backend is configured to the current working directory to enable faster debugging during development.

We use `WebIdentityTokenFile` and also set it in environment variables instead of setting the values in config, since the config values are cached in the state file and while refresh the newly set values are not picked up. This is a known issue with Pulumi. See the following issues for more details:

- https://github.com/pulumi/pulumi/issues/4981
- https://github.com/pulumi/pulumi-aws/issues/3149


### Destroy resources

To destroy the resources, run the program with the `destroy` flag.

```bash
go run . destroy
```

This will destroy the resources created by the program and remove the stack.
