package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	omsv1 "oms-contract/api/proto"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := omsv1.NewOMSClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fmt.Println("------------- Create Order -------------")
	r, err := c.CreateOrder(ctx, &omsv1.CreateOrderRequest{
		UserId:   12345,
		Symbol:   "BTCUSDT",
		Side:     omsv1.Side_SIDE_BUY,
		Type:     omsv1.OrderType_ORDER_TYPE_LIMIT,
		Price:    50000,
		Quantity: 1.5,
	})
	if err != nil {
		log.Fatalf("could not create order: %v", err)
	}
	fmt.Printf("Order Created: ID=%d Status=%s\n", r.OrderId, r.Status)

	time.Sleep(100 * time.Millisecond) // Wait for async processing if needed

	fmt.Println("------------- Get Position -------------")
	// Position won't be created until order is matched actually. 
	// But let's check if we get "not found" or empty.
	// Actually CreateOrder only creates an order. No trade execution in this test unless engine runs.
	// But let's try calling GetPosition to ensure connectivity.
	
	// Create a dummy position via a backdoor? No.
	// Just verify GetPosition returns error (Not Found) or success.
	
	p, err := c.GetPosition(ctx, &omsv1.GetPositionRequest{
		UserId: 12345,
		Symbol: "BTCUSDT",
	})
	if err != nil {
		fmt.Printf("GetPosition check: %v (Expected if no trades)\n", err)
	} else {
		fmt.Printf("Position: %+v\n", p)
	}
}
