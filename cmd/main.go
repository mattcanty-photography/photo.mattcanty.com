package main

import (
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"

	"github.com/matt.canty/photo.mattcanty.com/platform/internal/cdn"
	"github.com/matt.canty/photo.mattcanty.com/platform/internal/photos"
	"github.com/matt.canty/photo.mattcanty.com/platform/internal/site"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		photosResult, err := photos.CreatePhotosResources(ctx)
		if err != nil {
			return err
		}

		siteResult, err := site.CreateSiteResources(ctx, photosResult)
		if err != nil {
			return err
		}

		err = cdn.CreateCDN(
			ctx,
			photosResult,
			siteResult,
		)
		if err != nil {
			return err
		}

		return nil
	})
}
