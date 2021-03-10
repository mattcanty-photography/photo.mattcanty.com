package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func createPhotosResources(ctx *pulumi.Context) (pulumi.StringOutput, error) {
	bucket, err := s3.NewBucket(
		ctx,
		awsNamePrintf(ctx, "%s", "photos"),
		&s3.BucketArgs{})

	_, err = cloudfront.NewDistribution(
		ctx,
		awsNamePrintf(ctx, "%s", "photos"),
		&cloudfront.DistributionArgs{
			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					DomainName: bucket.BucketRegionalDomainName,
					OriginId:   bucket.ID().ToStringOutput(),
				},
			},
			Enabled:       pulumi.Bool(true),
			IsIpv6Enabled: pulumi.Bool(true),
			DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
				AllowedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
					pulumi.String("OPTIONS"),
				},
				CachedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},
				TargetOriginId: bucket.ID().ToStringOutput(),
				ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
					QueryString: pulumi.Bool(false),
					Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
						Forward: pulumi.String("none"),
					},
				},
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				MinTtl:               pulumi.Int(0),
				DefaultTtl:           pulumi.Int(3600),
				MaxTtl:               pulumi.Int(86400),
			},
			PriceClass: pulumi.String("PriceClass_100"),
			Restrictions: &cloudfront.DistributionRestrictionsArgs{
				GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
					RestrictionType: pulumi.String("none"),
				},
			},
			ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
				CloudfrontDefaultCertificate: pulumi.Bool(true),
			},
		},
	)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

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
		return pulumi.StringOutput{}, err
	}

	s3.NewBucketPolicy(
		ctx,
		awsNamePrintf(ctx, "%s", "photos"),
		&s3.BucketPolicyArgs{
			Bucket: bucket.ID(),
			Policy: tmpJSON,
		},
	)

	return bucket.Bucket, err
}
