package photos

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"

	"github.com/matt.canty/photo.mattcanty.com/platform/internal/helpers"
)

func CreatePhotosResources(ctx *pulumi.Context) (s3.Bucket, error) {
	bucket, err := s3.NewBucket(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photos"),
		&s3.BucketArgs{})

	tmpJSON := bucket.Arn.ApplyT(func(arn string) (string, error) {
		policyJSON, err := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect":    "Allow",
					"Principal": "*",
					"Action":    []string{"s3:GetObject"},
					"Resource": []string{
						fmt.Sprintf("%s/*", arn),
					},
				},
			},
		})
		if err != nil {
			return "", err
		}
		return string(policyJSON), nil
	})
	if err != nil {
		return s3.Bucket{}, err
	}

	s3.NewBucketPolicy(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photos"),
		&s3.BucketPolicyArgs{
			Bucket: bucket.ID(),
			Policy: tmpJSON,
		},
	)

	var doc helpers.AssumeRolePolicyDocument
	doc.Version = "2012-10-17"
	doc.Statement = []helpers.AssumeRolePolicyStatmentEntry{
		{
			Sid:    "",
			Effect: "Allow",
			Action: "sts:AssumeRole",
			Principal: helpers.AssumeRolePolicyStatmentEntryPrincipal{
				Service: "lambda.amazonaws.com",
			},
		},
	}

	assumeRolePolicy, err := json.Marshal(&doc)
	if err != nil {
		return s3.Bucket{}, err
	}

	lambdaRole, err := iam.NewRole(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photosapi-lambda"),
		&iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		},
	)
	if err != nil {
		return s3.Bucket{}, err
	}

	policyTmp := bucket.Bucket.ApplyString(func(bucketID string) (string, error) {
		policyStatement := []helpers.PolicyStatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"s3:Get*",
					"s3:Describe*",
					"s3:List*",
				},
				Resource: []string{
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketID),
					fmt.Sprintf("arn:aws:s3:::%s", bucketID),
				},
			},
			{
				Effect: "Allow",
				Action: []string{
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents",
				},
				Resource: []string{
					"arn:aws:logs:*:*:*",
				},
			},
			{
				Effect: "Allow",
				Action: []string{
					"xray:PutTraceSegments",
					"xray:PutTelemetryRecords",
					"xray:GetSamplingRules",
					"xray:GetSamplingTargets",
					"xray:GetSamplingStatisticSummaries",
				},
				Resource: []string{
					"*",
				},
			},
		}

		policyFormat, policyArgs, err := helpers.NewPolicyDocumentString(policyStatement...)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(policyFormat, policyArgs...), nil
	})

	lambdaRolePolicy, err := iam.NewRolePolicy(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photosapi-lambda"),
		&iam.RolePolicyArgs{
			Role:   lambdaRole.Name,
			Policy: policyTmp,
		},
	)
	if err != nil {
		return s3.Bucket{}, err
	}

	photosapiVersion := "v0.0.1-beta.rc1"

	remoteArchive := fmt.Sprintf(
		"https://github.com/mattcanty-photography/photosapi/releases/download/%s/photosapi_%s_linux_amd64.zip",
		photosapiVersion,
		photosapiVersion,
	)

	_, err = lambda.NewFunction(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photosapi"),
		&lambda.FunctionArgs{
			Handler: pulumi.String("handler"),
			Role:    lambdaRole.Arn,
			Runtime: pulumi.String("go1.x"),
			Code:    pulumi.NewRemoteArchive(remoteArchive),
			TracingConfig: lambda.FunctionTracingConfigArgs{
				Mode: pulumi.String("Active"),
			},
			Timeout: pulumi.Int(10),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"PHOTO_BUCKET_NAME": bucket.ID(),
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{lambdaRolePolicy}),
	)
	if err != nil {
		return s3.Bucket{}, err
	}

	return *bucket, err
}
