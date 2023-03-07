package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/dirien/pulumi-connector/internal/port"
	"github.com/dirien/pulumi-connector/internal/pulumi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	Port             int
	PortClientId     string
	PortClientSecret string
	PortBaseUrl      string

	portClient *port.Client
	pu         *pulumi.Pulumi
)

func main() {
	e := echo.New()
	flag.IntVar(&Port, "port", 8080, "Port for test HTTP server")
	PortClientId, _ = os.LookupEnv("PORT_CLIENT_ID")
	if PortClientId == "" {
		e.Logger.Fatal("PORT_CLIENT_ID is not set")
	}
	PortClientSecret, _ = os.LookupEnv("PORT_CLIENT_SECRET")
	if PortClientSecret == "" {
		e.Logger.Fatal("PORT_CLIENT_SECRET is not set")
	}
	PortBaseUrl, _ = os.LookupEnv("PORT_BASE_URL")
	if PortBaseUrl == "" {
		PortBaseUrl = "https://api.getport.io"
	}
	if os.Getenv("PULUMI_ACCESS_TOKEN") == "" {
		e.Logger.Fatal("PULUMI_ACCESS_TOKEN is set")
	}
	flag.Parse()

	e.Use(middleware.Logger())

	e.Logger.Info("Authenticating with Port")
	portClient = port.New(PortBaseUrl)
	_, err := portClient.Authenticate(context.Background(), PortClientId, PortClientSecret)
	if err != nil {
		e.Logger.Fatalf("failed to authenticate with Port: %v", err)
	}
	pu = pulumi.NewPulumi(&e.Logger)

	e.POST("/", func(c echo.Context) error {
		err = actionHandler(c)
		if err != nil {
			e.Logger.Errorf("%v", err)
			c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			return nil
		}
		c.JSON(http.StatusOK, echo.Map{"message": "ok"})
		return nil
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", Port)))
}

func actionHandler(c echo.Context) (err error) {
	body := port.ActionBody{}
	err = c.Bind(&body)
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch body.Payload.Action.Trigger {
	case "CREATE", "DAY-2":
		c.Logger().Infof("Running create action: %s", body.Payload.Action.Identifier)
		err = pu.Up(ctx, &body)
	case "DELETE":
		c.Logger().Infof("Running delete action: %s", body.Payload.Action.Identifier)
		err = pu.Destroy(ctx, &body)
	default:
		return fmt.Errorf("unknown action: %s", body.Payload.Action.Identifier)
	}
	if err != nil {
		portClient.PatchActionRun(ctx, body.Context.RunID, port.ActionStatusFailure)
		return err
	}
	err = portClient.PatchActionRun(ctx, body.Context.RunID, port.ActionStatusSuccess)
	if err != nil {
		return err
	}
	return nil
}
