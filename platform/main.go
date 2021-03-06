package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		photosBucketName, err := createPhotosResources(ctx)
		if err != nil {
			return err
		}

		err = createSiteResources(ctx, photosBucketName)
		if err != nil {
			return err
		}

		return nil
	})
}

func awsNamePrintf(ctx *pulumi.Context, format string, a ...interface{}) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	ctxString := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())
	fmtString := fmt.Sprintf(format, a...)

	outString := strings.Join([]string{ctxString, fmtString}, "-")
	outString = reg.ReplaceAllString(outString, "-")
	outString = strings.Trim(outString, "-")

	return outString
}
