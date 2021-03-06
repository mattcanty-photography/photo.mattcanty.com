package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func createSiteResources(ctx *pulumi.Context, photosBucketName string) error {
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

	policyStatement := []policyStatementEntry{
		{
			Effect: "Allow",
			Action: []string{
				"s3:Get*",
				"s3:Describe*",
				"s3:List*",
			},
			Resource: []string{
				fmt.Sprintf("arn:aws:s3:::%s/*", photosBucketName),
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
		return err
	}

	lambdaRole, err := iam.NewRole(
		ctx,
		awsNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		})

	lambdaRolePolicy, err := iam.NewRolePolicy(
		ctx,
		awsNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RolePolicyArgs{
			Role:   lambdaRole.Name,
			Policy: pulumi.Sprintf(policyFormat, policyArgs...),
		})

	function, err := lambda.NewFunction(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&lambda.FunctionArgs{
			Handler: pulumi.String("www"),
			Role:    lambdaRole.Arn,
			Runtime: pulumi.String("go1.x"),
			Code:    pulumi.NewFileArchive("../www/.build/www.zip"),
			TracingConfig: lambda.FunctionTracingConfigArgs{
				Mode: pulumi.String("Active"),
			},
			Timeout: pulumi.Int(10),
		},
		pulumi.DependsOn([]pulumi.Resource{lambdaRolePolicy}),
	)

	api, err := apigatewayv2.NewApi(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&apigatewayv2.ApiArgs{
			ProtocolType: pulumi.String("HTTP"),
		})
	if err != nil {
		return err
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
		return err
	}

	_, err = apigatewayv2.NewRoute(
		ctx,
		"root",
		&apigatewayv2.RouteArgs{
			ApiId:    api.ID(),
			RouteKey: pulumi.String("$default"),
			Target:   pulumi.Sprintf("integrations/%s", integration.ID()),
		},
	)
	if err != nil {
		return err
	}

	_, err = lambda.NewPermission(
		ctx,
		awsNamePrintf(ctx, "%s", "site"),
		&lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("arn:aws:execute-api:eu-west-2:171269869827:%s/*/$default", api.ID()),
		}, pulumi.DependsOn([]pulumi.Resource{api, function}))
	if err != nil {
		return err
	}

	_, err = apigatewayv2.NewStage(ctx, ctx.Stack(), &apigatewayv2.StageArgs{
		ApiId:      api.ID(),
		AutoDeploy: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	return nil
}
