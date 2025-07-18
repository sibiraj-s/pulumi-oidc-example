# Pulumi OIDC Example

This is an example of how to use OIDC with Pulumi automation api to assume a role in AWS and create resources.

## Prerequisites

- Go (check the go version in the `go.mod` file)
- Pulumi v3.183.0+

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

### Destroy resources

To destroy the resources, run the program with the `destroy` flag.

```bash
go run . destroy
```

This will destroy the resources created by the program and remove the stack.
