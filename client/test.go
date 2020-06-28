package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func main() {
	ctx := context.Background()

	// Step 1: Create a client
	clientCfg := client.NewConfiguration(
		"http://localhost:8080",
		"ckb-rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)

	client := client.NewAPIClient(clientCfg)

	var index int64 = 1381633

	// Step 11: Fetch the current block
	block, rosettaErr, err := client.BlockAPI.Block(
		ctx,
		&types.BlockRequest{
			NetworkIdentifier: &types.NetworkIdentifier{
				Blockchain: "CKB",
				Network:    "Mainnet",
			},
			BlockIdentifier: &types.PartialBlockIdentifier{
				Index: &index,
			},
		},
	)
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Step 12: Print the block
	prettyBlock, err := json.MarshalIndent(block.Block, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Current Block: %s\n", string(prettyBlock))

}
