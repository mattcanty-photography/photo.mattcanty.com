package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func createCDN(
	ctx *pulumi.Context,
	photosBucketDomainName pulumi.StringOutput,
	apiGatewayDomainName pulumi.StringOutput,
	apiGatewayStageName pulumi.StringOutput,
) error {
	current, err := aws.GetCallerIdentity(ctx, nil, nil)
	if err != nil {
		return err
	}

	_, err = cloudfront.NewDistribution(
		ctx,
		awsNamePrintf(ctx, "%s", "default"),
		&cloudfront.DistributionArgs{
			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					OriginId:   pulumi.String("photos"),
					DomainName: photosBucketDomainName,
				},
				cloudfront.DistributionOriginArgs{
					OriginId:   pulumi.String("www"),
					DomainName: apiGatewayDomainName,
					OriginPath: pulumi.Sprintf("/%s", apiGatewayStageName),
					CustomOriginConfig: cloudfront.DistributionOriginCustomOriginConfigArgs{
						HttpPort:  pulumi.Int(80),
						HttpsPort: pulumi.Int(443),
						OriginSslProtocols: pulumi.StringArray{
							pulumi.String("TLSv1.2"),
							pulumi.String("TLSv1.1"),
							pulumi.String("TLSv1"),
							pulumi.String("SSLv3"),
						},
						OriginProtocolPolicy: pulumi.String("https-only"),
					},
				},
			},
			Enabled:       pulumi.Bool(true),
			IsIpv6Enabled: pulumi.Bool(true),
			DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
				TargetOriginId: pulumi.String("www"),
				AllowedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
					pulumi.String("OPTIONS"),
				},
				CachedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},
				ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
					QueryString: pulumi.Bool(false),
					Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
						Forward: pulumi.String("none"),
					},
				},
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				MinTtl:               pulumi.Int(0),
				DefaultTtl:           pulumi.Int(0),
				MaxTtl:               pulumi.Int(0),
			},
			OrderedCacheBehaviors: cloudfront.DistributionOrderedCacheBehaviorArray{
				cloudfront.DistributionOrderedCacheBehaviorArgs{
					TargetOriginId: pulumi.String("photos"),
					PathPattern:    pulumi.String("/portfolios/*"),
					AllowedMethods: pulumi.StringArray{
						pulumi.String("GET"),
						pulumi.String("HEAD"),
						pulumi.String("OPTIONS"),
					},
					CachedMethods: pulumi.StringArray{
						pulumi.String("GET"),
						pulumi.String("HEAD"),
					},
					ForwardedValues: cloudfront.DistributionOrderedCacheBehaviorForwardedValuesArgs{
						QueryString: pulumi.Bool(false),
						Cookies: &cloudfront.DistributionOrderedCacheBehaviorForwardedValuesCookiesArgs{
							Forward: pulumi.String("none"),
						},
					},
					ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
					MinTtl:               pulumi.Int(3600),
					DefaultTtl:           pulumi.Int(86400),
					MaxTtl:               pulumi.Int(86400),
				},
			},
			PriceClass: pulumi.String("PriceClass_100"),
			Restrictions: &cloudfront.DistributionRestrictionsArgs{
				GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
					RestrictionType: pulumi.String("none"),
				},
			},
			ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
				AcmCertificateArn: pulumi.Sprintf(
					"arn:aws:acm:us-east-1:%s:certificate/d10760d0-8de4-40c2-80d4-2d4211982d17",
					current.AccountId,
				),
				SslSupportMethod: pulumi.String("sni-only"),
			},
			Aliases: pulumi.StringArray{
				pulumi.String("photo.mattcanty.com"),
			},
		},
	)
	if err != nil {
		return err
	}

	return err
}
