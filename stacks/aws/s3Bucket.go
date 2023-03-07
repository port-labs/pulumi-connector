package aws

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"strings"

	port_ "github.com/dirien/pulumi-port-labs/sdk/go/port"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func S3Bucket() pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		tags := pulumi.StringMap{}
		var tagsFromConfig = config.Get(ctx, "tags")
		if tagsFromConfig != "" {
			tagKeyPair := strings.Split(tagsFromConfig, ",")
			for _, v := range tagKeyPair {
				tag := strings.Split(v, "=")
				if len(tag) == 2 {
					tags[tag[0]] = pulumi.String(tag[1])
				}
			}
		} else {
			tags = pulumi.StringMap{}
		}

		acl := pulumi.String("private")
		fmt.Println(len(config.Get(ctx, "bucket_acl")))
		if len(config.Get(ctx, "bucket_acl")) != 0 {
			acl = pulumi.String(config.Get(ctx, "bucket_acl"))
		}

		bucket, err := s3.NewBucket(ctx, "bucket", &s3.BucketArgs{
			Acl:    acl,
			Bucket: pulumi.String(config.Get(ctx, "bucket_name")),
			Tags:   tags,
		})
		if err != nil {
			return err
		}
		_, err = port_.NewEntity(ctx, "entity", &port_.EntityArgs{
			RunId:      pulumi.String(config.Get(ctx, "run_id")),
			Blueprint:  pulumi.String(config.Get(ctx, "blueprint")),
			Identifier: pulumi.String(config.Get(ctx, "entity_identifier")),
			Title:      pulumi.Sprintf("Bucket %s", bucket.Bucket),
			Properties: port_.EntityPropertyArray{
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("bucket_name"),
					Value: bucket.Bucket,
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("bucket_acl"),
					Value: bucket.Acl,
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("tags"),
					Value: pulumi.JSONMarshal(bucket.Tags),
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("url"),
					Value: pulumi.Sprintf("https://s3.console.aws.amazon.com/s3/buckets/%s", bucket.Bucket),
				},
			},
		})
		if err != nil {
			return err
		}
		return nil
	}
}
