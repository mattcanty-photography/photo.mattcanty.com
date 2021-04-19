package main

import (
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"

	"github.com/matt.canty/photo.mattcanty.com/platform/internal/cdn"
	"github.com/matt.canty/photo.mattcanty.com/platform/internal/photos"
	"github.com/matt.canty/photo.mattcanty.com/platform/internal/site"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		photosBucket, err := photos.CreatePhotosResources(ctx)
		if err != nil {
			return err
		}

		apiGatewayDomainName, apiGatewayStageName, err := site.CreateSiteResources(ctx, photosBucket.Bucket)
		if err != nil {
			return err
		}

		err = cdn.CreateCDN(
			ctx,
			photosBucket.BucketRegionalDomainName,
			apiGatewayDomainName,
			apiGatewayStageName,
		)
		if err != nil {
			return err
		}

		return nil
	})
}
