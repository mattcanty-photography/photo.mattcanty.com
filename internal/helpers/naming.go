package helpers

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func AWSNamePrintf(ctx *pulumi.Context, format string, a ...interface{}) string {
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
