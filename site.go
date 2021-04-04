package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func createSiteResources(
	ctx *pulumi.Context,
	bucketName pulumi.StringOutput) (pulumi.StringOutput, pulumi.StringOutput, error) {
	var doc assumeRolePolicyDocument
	doc.Version = "2012-10-17"
	doc.Statement = []assumeRolePolicyStatmentEntry{
		{
			Sid:    "",
			Effect: "Allow",
			Action: "sts:AssumeRole",
			Principal: assumeRolePolicyStatmentEntryPrincipal{
				Service: "lambda.amazonaws.com",
			},
		},
	}

	assumeRolePolicy, err := json.Marshal(&doc)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	lambdaRole, err := iam.NewRole(
		ctx,
		awsNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	policyTmp := bucketName.ApplyString(func(bucketID string) (string, error) {
		policyStatement := []policyStatementEntry{
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

		policyFormat, policyArgs, err := newPolicyDocumentString(policyStatement...)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(policyFormat, policyArgs...), nil
	})

	lambdaRolePolicy, err := iam.NewRolePolicy(
		ctx,
		awsNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RolePolicyArgs{
			Role:   lambdaRole.Name,
			Policy: policyTmp,
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	websiteVersion := "v0.0.7"

	remoteArchive := fmt.Sprintf(
		"https://github.com/mattcanty-photography/website/releases/download/%s/website_%s_linux_amd64.zip",
		websiteVersion,
		websiteVersion,
	)

	function, err := lambda.NewFunction(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
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
					"PHOTO_BUCKET_NAME": bucketName,
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{lambdaRolePolicy}),
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	api, err := apigatewayv2.NewApi(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&apigatewayv2.ApiArgs{
			ProtocolType: pulumi.String("HTTP"),
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	integration, err := apigatewayv2.NewIntegration(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&apigatewayv2.IntegrationArgs{
			ApiId:                api.ID(),
			IntegrationType:      pulumi.String("AWS_PROXY"),
			ConnectionType:       pulumi.String("INTERNET"),
			IntegrationMethod:    pulumi.String("POST"),
			IntegrationUri:       function.InvokeArn,
			PayloadFormatVersion: pulumi.String("2.0"),
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	defaultRoute, err := apigatewayv2.NewRoute(
		ctx,
		"root",
		&apigatewayv2.RouteArgs{
			ApiId:    api.ID(),
			RouteKey: pulumi.String("$default"),
			Target:   pulumi.Sprintf("integrations/%s", integration.ID()),
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	region, err := aws.GetRegion(ctx, nil, nil)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	current, err := aws.GetCallerIdentity(ctx, nil, nil)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	_, err = lambda.NewPermission(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf(
				"arn:aws:execute-api:%s:%s:%s/*/%s",
				region.Name,
				current.AccountId,
				api.ID(),
				defaultRoute.RouteKey,
			),
		}, pulumi.DependsOn([]pulumi.Resource{api, function}))
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	stage, err := apigatewayv2.NewStage(ctx, ctx.Stack(), &apigatewayv2.StageArgs{
		ApiId:      api.ID(),
		AutoDeploy: pulumi.Bool(true),
	})
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	apiDomainName := pulumi.Sprintf(
		"%s.execute-api.%s.amazonaws.com",
		api.ID(),
		region.Name,
	)

	return apiDomainName, stage.Name, nil
}
