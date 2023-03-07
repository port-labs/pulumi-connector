package pulumi

import (
	"context"
	"fmt"
	"strings"

	"github.com/dirien/pulumi-connector/internal/port"
	"github.com/dirien/pulumi-connector/templates"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/samber/lo"
)

type Pulumi struct {
	logger *echo.Logger
}

func genUUID(len int) string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)[:len]
}

func getStateKey(actionBody *port.ActionBody) string {
	if actionBody.Context.Entity != "" {
		return actionBody.Context.Entity
	}
	return "e_" + genUUID(16)
}

func NewPulumi(logger *echo.Logger) *Pulumi {
	return &Pulumi{
		logger: logger,
	}
}

func (p *Pulumi) Destroy(ctx context.Context, actionBody *port.ActionBody) error {
	stateKey := actionBody.Context.Entity
	props := lo.Assign(map[string]any{}, actionBody.Payload.Entity.Properties, map[string]any{
		"entity_identifier": stateKey,
		"blueprint":         actionBody.Context.Blueprint,
		"run_id":            actionBody.Context.RunID,
	})
	program, err := templates.Program(ctx, props)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	stdoutStreamer := optdestroy.ProgressStreams((*p.logger).Output())
	_, err = program.Stack.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("error refreshing stack: %v", err)
	}
	_, err = program.Stack.Destroy(ctx, stdoutStreamer)
	if err != nil {
		return fmt.Errorf("error destroying stack: %v", err)
	}
	err = program.Stack.Workspace().RemoveStack(ctx, program.StackName)
	if err != nil {
		return fmt.Errorf("failed to remove stack: %v", err)
	}
	return nil
}

func (p *Pulumi) Up(ctx context.Context, actionBody *port.ActionBody) error {
	stateKey := getStateKey(actionBody)
	props := lo.Assign(map[string]any{}, actionBody.Payload.Entity.Properties, actionBody.Payload.Properties, map[string]any{
		"entity_identifier": stateKey,
		"blueprint":         actionBody.Context.Blueprint,
		"run_id":            actionBody.Context.RunID,
	})

	program, err := templates.Program(ctx, props)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	stdoutStreamer := optup.ProgressStreams((*p.logger).Output())
	_, err = program.Stack.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("error refreshing stack: %v", err)
	}
	_, err = program.Stack.Up(ctx, stdoutStreamer)
	if err != nil {
		_, err = program.Stack.Destroy(ctx)
		if err != nil {
			return fmt.Errorf("failed to destroy stack: %v", err)
		}
		err = program.Stack.Workspace().RemoveStack(ctx, program.StackName)
		if err != nil {
			return fmt.Errorf("failed to remove stack: %v", err)
		}
		return fmt.Errorf("failed update: %v", err)
	}
	return nil
}
