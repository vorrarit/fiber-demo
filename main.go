package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var config Config = ReadConfig()
var tracer = otel.Tracer(config.Application.Name)
var tp trace.TracerProvider
var ctx context.Context = context.Background()

func main() {
	app := fiber.New()

	if config.Application.Otel.Enable {
		tp, err := setupTracing(ctx, config.Application.Name, config.Application.Otel.Grpc_Url)
		if err != nil {
			panic(err)
		}
		defer func() {
			err := tp.Shutdown(ctx)
			fmt.Println("error when shutting down tracingProvider. err: ", err)
		}()
		app.Use(otelfiber.Middleware(otelfiber.WithSpanNameFormatter(func(ctx *fiber.Ctx) string {
			return fmt.Sprintf("%s %s", ctx.Method(), ctx.Route().Path)
		})))
	}

	app.Post("/echo", echo)
	app.Post("/serviceb", serviceb)

	err := app.Listen(fmt.Sprintf(":%d", config.Application.Port))
	if err != nil {
		fmt.Printf("Error starting server - %v+", err)
	}

}

func echo(c *fiber.Ctx) error {
	log := NewSlog(c.UserContext())

	headers := c.GetReqHeaders()
	log.Debug("### HEADERS ###")

	for headerName, headerValue := range headers {
		log.Debug(fmt.Sprintf("%s %s", headerName, headerValue))
		c.Set(fmt.Sprintf("REQ_%s", headerName), strings.Join(headerValue[:], " "))
	}

	log.Debug("### BODY ###")
	body := string(c.Body()[:])
	log.Debug(body)
	return c.Send([]byte(body))
}

func serviceb(c *fiber.Ctx) error {
	log := NewSlog(c.UserContext())

	headers := c.GetReqHeaders()
	log.Debug("### HEADERS ###")
	for headerName, headerValue := range headers {
		log.Debug(fmt.Sprintf("%s %s", headerName, headerValue))
	}

	log.Debug("### BODY ###")
	body := string(c.Body()[:])
	log.Debug(body)

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	byteBody := []byte(body)
	bodyReader := bytes.NewReader(byteBody)
	cli_request, _ := http.NewRequestWithContext(c.UserContext(), http.MethodPost, config.ServiceB.Url, bodyReader)
	cli_response, _ := client.Do(cli_request)
	// cli_response, _ := otelhttp.Post(c.UserContext(), "http://localhost:8080/echo", "text/plain", bodyReader)
	cli_response_body, _ := io.ReadAll(cli_response.Body)

	for key, values := range cli_response.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	return c.Send(cli_response_body)
}
