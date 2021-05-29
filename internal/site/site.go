package site

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"

	"github.com/matt.canty/photo.mattcanty.com/platform/internal/helpers"
	"github.com/matt.canty/photo.mattcanty.com/platform/internal/photos"
)

type SiteResult struct {
	DomainName pulumi.StringOutput
	StageName  pulumi.StringOutput
}

func CreateSiteResources(
	ctx *pulumi.Context,
	photosResult photos.PhotosResult) (SiteResult, error) {
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
		return SiteResult{}, err
	}

	lambdaRole, err := iam.NewRole(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	policyTmp := photosResult.Bucket.Bucket.ApplyString(func(bucketID string) (string, error) {
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
		helpers.AWSNamePrintf(ctx, "%s", "site-lambda"),
		&iam.RolePolicyArgs{
			Role:   lambdaRole.Name,
			Policy: policyTmp,
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	websiteVersion := "v0.0.8-beta.rc1"

	remoteArchive := fmt.Sprintf(
		"https://github.com/mattcanty-photography/website/releases/download/%s/website_%s_linux_amd64.zip",
		websiteVersion,
		websiteVersion,
	)

	function, err := lambda.NewFunction(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site"),
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
					"PHOTO_BUCKET_NAME": photosResult.Bucket.Bucket,
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{lambdaRolePolicy}),
	)
	if err != nil {
		return SiteResult{}, err
	}

	api, err := apigatewayv2.NewApi(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site"),
		&apigatewayv2.ApiArgs{
			ProtocolType: pulumi.String("HTTP"),
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	defaultIntegration, err := apigatewayv2.NewIntegration(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site"),
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
		return SiteResult{}, err
	}

	defaultRoute, err := apigatewayv2.NewRoute(
		ctx,
		"root",
		&apigatewayv2.RouteArgs{
			ApiId:    api.ID(),
			RouteKey: pulumi.String("$default"),
			Target:   pulumi.Sprintf("integrations/%s", defaultIntegration.ID()),
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	photosIntegration, err := apigatewayv2.NewIntegration(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photos"),
		&apigatewayv2.IntegrationArgs{
			ApiId:                api.ID(),
			IntegrationType:      pulumi.String("AWS_PROXY"),
			ConnectionType:       pulumi.String("INTERNET"),
			IntegrationMethod:    pulumi.String("POST"),
			IntegrationUri:       photosResult.S3Function.InvokeArn,
			PayloadFormatVersion: pulumi.String("2.0"),
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	_, err = apigatewayv2.NewRoute(
		ctx,
		"photos",
		&apigatewayv2.RouteArgs{
			ApiId:    api.ID(),
			RouteKey: pulumi.String("ANY /photo/{proxy+}"),
			Target:   pulumi.Sprintf("integrations/%s", photosIntegration.ID()),
		},
	)
	if err != nil {
		return SiteResult{}, err
	}

	region, err := aws.GetRegion(ctx, nil, nil)
	if err != nil {
		return SiteResult{}, err
	}

	current, err := aws.GetCallerIdentity(ctx, nil, nil)
	if err != nil {
		return SiteResult{}, err
	}

	_, err = lambda.NewPermission(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site"),
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
		return SiteResult{}, err
	}

	_, err = lambda.NewPermission(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "photos"),
		&lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  photosResult.S3Function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf(
				"arn:aws:execute-api:%s:%s:%s/*",
				region.Name,
				current.AccountId,
				api.ID(),
			),
		}, pulumi.DependsOn([]pulumi.Resource{api, photosResult.S3Function}))
	if err != nil {
		return SiteResult{}, err
	}

	logGroup, err := cloudwatch.NewLogGroup(
		ctx,
		helpers.AWSNamePrintf(ctx, "%s", "site"),
		&cloudwatch.LogGroupArgs{},
	)
	if err != nil {
		return SiteResult{}, err
	}

	stage, err := apigatewayv2.NewStage(ctx, ctx.Stack(), &apigatewayv2.StageArgs{
		ApiId:      api.ID(),
		AutoDeploy: pulumi.Bool(true),
		AccessLogSettings: apigatewayv2.StageAccessLogSettingsArgs{
			Format:         pulumi.String("$context.identity.sourceIp - - [$context.requestTime] \"$context.httpMethod $context.routeKey $context.protocol\" $context.status $context.responseLength $context.requestId $context.integrationErrorMessage"),
			DestinationArn: pulumi.Sprintf("arn:aws:logs:%s:%s:log-group:%s:*", region.Name, current.AccountId, logGroup.Name),
		},
	})
	if err != nil {
		return SiteResult{}, err
	}

	apiDomainName := pulumi.Sprintf(
		"%s.execute-api.%s.amazonaws.com",
		api.ID(),
		region.Name,
	)

	return SiteResult{
		DomainName: apiDomainName,
		StageName:  stage.Name,
	}, nil
}
