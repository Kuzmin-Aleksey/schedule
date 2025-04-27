package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	schedulev1 "schedule/gen-proto"
)

func main() {
	cc, err := grpc.DialContext(context.Background(),
		"127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal(err)
	}
	client := schedulev1.NewScheduleClient(cc)

	md := metadata.New(map[string]string{
		"TZ": "+02:00",
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	reply, err := client.GetNextTakings(ctx, &schedulev1.GetNextTakingsRequest{
		UserId: 123,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", reply.Items)
}
