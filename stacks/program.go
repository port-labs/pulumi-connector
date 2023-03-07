package stacks

import (
	"context"
	"fmt"
	"github.com/dirien/pulumi-connector/stacks/civo"
	"github.com/labstack/echo/v4"
	"reflect"

	"github.com/dirien/pulumi-connector/stacks/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ProgramResult struct {
	Stack       auto.Stack
	ProjectName string
	StackName   string
}

func Program(ctx context.Context, props map[string]any, log echo.Logger) (ProgramResult, error) {
	// default values should not be used
	projectName := "port-labs-test"
	stackName := "test"

	var program pulumi.RunFunc
	if props["blueprint"] == "s3_bucket" {
		projectName = fmt.Sprintf("bucket_%s", props["entity_identifier"])
		stackName = fmt.Sprintf("%s", props["entity_identifier"])
		program = aws.S3Bucket()
	} else if props["blueprint"] == "civo_cluster" {
		projectName = fmt.Sprintf("civo_cluster_%s", props["entity_identifier"])
		stackName = fmt.Sprintf("%s", props["entity_identifier"])
		program = civo.KubernetesCluster()
	} else {
		return ProgramResult{
			Stack:       auto.Stack{},
			ProjectName: projectName,
			StackName:   stackName,
		}, fmt.Errorf("unknown blueprint: '%s'", props["blueprint"])
	}

	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, program)
	if err != nil {
		return ProgramResult{
			Stack:       s,
			ProjectName: projectName,
			StackName:   stackName,
		}, err
	}

	w := s.Workspace()

	log.Info("Installing AWS Plugin")
	err = w.InstallPlugin(ctx, "aws", "v5.30.0")
	if err != nil {
		return ProgramResult{
			Stack:       s,
			ProjectName: projectName,
			StackName:   stackName,
		}, fmt.Errorf("error installing AWS resource plugin: %v", err)
	}
	log.Info("Installing Civo Plugin")
	err = w.InstallPlugin(ctx, "civo", "v2.3.3")
	if err != nil {
		return ProgramResult{
			Stack:       s,
			ProjectName: projectName,
			StackName:   stackName,
		}, fmt.Errorf("error installing Civo resource plugin: %v", err)
	}
	log.Info("Installing Port Plugin")
	err = w.InstallPluginFromServer(ctx, "port", "v0.8.3", "https://github.com/dirien/pulumi-port-labs/releases/download/v0.8.3")
	if err != nil {
		return ProgramResult{
			Stack:       s,
			ProjectName: projectName,
			StackName:   stackName,
		}, fmt.Errorf("error installing Port Labs resource plugin: %v", err)
	}
	region := props["region"].(string)
	s.SetConfig(ctx, "aws:region", auto.ConfigValue{
		Value: region,
	})
	s.SetConfig(ctx, "civo:region", auto.ConfigValue{
		Value: region,
	})

	for s2, a := range props {
		if s2 == "region" {
			continue
		}
		if reflect.TypeOf(a).Kind() == reflect.String {
			s.SetConfig(ctx, s2, auto.ConfigValue{
				Value: a.(string),
			})
		} else if reflect.TypeOf(a) == reflect.TypeOf(map[string]interface{}{}) {
			// Currently Pulumi Automation API does not support maps as config values
			var concat string
			for s3, b := range a.(map[string]interface{}) {
				concat = fmt.Sprintf("%s%s=%s,", concat, s3, b)
			}
			s.SetConfig(ctx, s2, auto.ConfigValue{})
		}
	}

	return ProgramResult{
		Stack:       s,
		ProjectName: projectName,
		StackName:   stackName,
	}, nil

}
