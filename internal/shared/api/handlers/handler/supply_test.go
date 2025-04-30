package handler

import (
	"context"
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc"
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"
)

func TestMe(t *testing.T) {
	tmClient, err := client.NewClientFromNode("https://babylon.nodes.guru/api")
	require.NoError(t, err)



	// Create a bank query client
	bankClient := banktypes.NewQueryClient(conn)

	// Query the total supply
	res, err := bankClient.TotalSupply(context.Background(), &banktypes.QueryTotalSupplyRequest{})
	if err != nil {
		log.Fatalf("failed to query total supply: %v", err)
	}

	// Print the total supply
	for _, coin := range res.Supply {
		fmt.Printf("Denom: %s, Amount: %s\n", coin.Denom, coin.Amount.String())
	}
}
