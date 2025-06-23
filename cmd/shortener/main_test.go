package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/darkseear/shortener/internal/proto"
)

func TestApp_Run_GRPCServer(t *testing.T) {

	grpServerAddr := "localhost:9090"
	testTimeout := 5 * time.Second
	testUrl := "http://example.com"

	conn, err := grpc.NewClient(grpServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewSortenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	respShort, err := client.Shorten(ctx, &proto.ShortenRequest{
		Url: testUrl,
	})
	if err != nil {
		t.Logf("Error resp short: %v", err)
	}
	t.Logf("ResponseShort: %v", respShort)
	respAdd, err := client.AddURL(ctx, &proto.AddURLRequest{
		Url: testUrl,
	})

	if err != nil {
		t.Logf("Error resp add: %v", err)
	}
	t.Logf("ResponseAdd: %v", respAdd)

	require.NoError(t, err)
	require.NotEmpty(t, respShort.ShortUrl)
}
