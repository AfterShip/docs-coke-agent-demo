package main

import (
	"context"
	"crypto/tls"
	platform_api_v2 "github.com/AfterShip/connectors-sdk-go/v2"
	"github.com/AfterShip/connectors-sdk-go/v2/orders"
	"github.com/AfterShip/gopkg/api/client"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/go-playground/validator/v10"
	"golang.org/x/time/rate"
	"log"
	"os"
	"time"
)

type GetOrderByIDArg struct {
	ID string `json:"id" description:"The ID of the order to retrieve"`
}

type AfterShipConnectorSDKTool struct {
	client *platform_api_v2.PlatformV2Client
}

func newAfterShipConnectorSDKTool() []tool.BaseTool {
	connectorTool := &AfterShipConnectorSDKTool{
		client: newConnectorClient(),
	}

	getOrderByIdTool, _ := utils.InferTool[GetOrderByIDArg, orders.ModelsResponseOrder](
		"get_order_by_id",
		"Get an order by its ID from AfterShip",
		connectorTool.GetOrderByID)
	return []tool.BaseTool{getOrderByIdTool}
}

func (t *AfterShipConnectorSDKTool) GetOrderByID(ctx context.Context, arg GetOrderByIDArg) (orders.ModelsResponseOrder, error) {
	log.Printf("Starting tool.calling [GetOrderByID], order_id: %s\n", arg.ID)
	// This function would typically interact with the AfterShip API to retrieve an order by its ID.
	// For now, we return a dummy order.
	orderSrv := orders.NewOrdersSvc(t.client)
	orderResp, getErr := orderSrv.GetOrdersByID(ctx, arg.ID, orders.GetOrdersByIDParams{})
	if getErr != nil {
		return orders.ModelsResponseOrder{}, getErr
	}

	if orderResp == nil {
		return orders.ModelsResponseOrder{}, nil
	}

	return *orderResp.Data, nil
}

func newConnectorClient() *platform_api_v2.PlatformV2Client {
	amAPIKey := os.Getenv("AM_API_KEY")
	if len(amAPIKey) == 0 {
		panic("AM_API_KEY environment variable is not set")
	}

	restyClient := client.New(nil)
	restyClient.SetHeader("am-api-key", amAPIKey)
	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//url := "https://platform.automizelyapi.io/connectors/v2"

	// this means every 10ms, add 1 token to bucket. equal to 100 qps
	limiter := rate.NewLimiter(rate.Every(10*time.Millisecond), 1)
	validate := validator.New()

	return platform_api_v2.NewPlatformV2ClientWithoutUrl(restyClient, limiter, validate)
}
