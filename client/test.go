package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	// serverURL is the URL of a Rosetta Server.
	serverURL = "http://localhost:8117/rosetta"

	// agent is the user-agent on requests to the
	// Rosetta Server.
	agent = "ckb-rosetta-sdk-go"

	// defaultTimeout is the default timeout for
	// HTTP requests.
	defaultTimeout = 10 * time.Second
)

func main() {
	ctx := context.Background()

	// Step 1: Create a client
	clientCfg := client.NewConfiguration(
		serverURL,
		agent,
		&http.Client{
			Timeout: defaultTimeout,
		},
	)

	client := client.NewAPIClient(clientCfg)

	// Step 2: Get all available networks
	networkList, rosettaErr, err := client.NetworkAPI.NetworkList(
		ctx,
		&types.MetadataRequest{},
	)
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	if len(networkList.NetworkIdentifiers) == 0 {
		log.Fatal("no available networks")
	}

	primaryNetwork := networkList.NetworkIdentifiers[0]

	// Step 3: Print the primary network
	prettyPrimaryNetwork, err := json.MarshalIndent(primaryNetwork, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Primary Network: %s\n", string(prettyPrimaryNetwork))

	// Step 4: Fetch the network status
	networkStatus, rosettaErr, err := client.NetworkAPI.NetworkStatus(
		ctx,
		&types.NetworkRequest{
			NetworkIdentifier: primaryNetwork,
		},
	)
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Step 5: Print the response
	prettyNetworkStatus, err := json.MarshalIndent(networkStatus, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Network Status: %s\n", string(prettyNetworkStatus))

	// Step 6: Assert the response is valid
	err = asserter.NetworkStatusResponse(networkStatus)
	if err != nil {
		log.Fatalf("Assertion Error: %s\n", err.Error())
	}

	// Step 7: Fetch the network options
	networkOptions, rosettaErr, err := client.NetworkAPI.NetworkOptions(
		ctx,
		&types.NetworkRequest{
			NetworkIdentifier: primaryNetwork,
		},
	)
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Step 8: Print the response
	prettyNetworkOptions, err := json.MarshalIndent(networkOptions, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Network Options: %s\n", string(prettyNetworkOptions))

	// Step 9: Assert the response is valid
	err = asserter.NetworkOptionsResponse(networkOptions)
	if err != nil {
		log.Fatalf("Assertion Error: %s\n", err.Error())
	}

	// Step 10: Create an asserter using the retrieved NetworkStatus and
	// NetworkOptions.
	//
	// This will be used later to assert that a fetched block is
	// valid.
	asserter, err := asserter.NewClientWithResponses(
		primaryNetwork,
		networkStatus,
		networkOptions,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Step 11: Fetch the current block
	block, rosettaErr, err := client.BlockAPI.Block(
		ctx,
		&types.BlockRequest{
			NetworkIdentifier: primaryNetwork,
			BlockIdentifier: types.ConstructPartialBlockIdentifier(
				networkStatus.CurrentBlockIdentifier,
			),
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

	// Step 13: Assert the block response is valid
	//
	// It is important to note that this only ensures
	// required fields are populated and that operations
	// in the block only use types and statuses that were
	// provided in the networkStatusResponse. To run more
	// intensive validation, use the Rosetta Validator. It
	// can be found at: https://github.com/coinbase/rosetta-validator
	err = asserter.Block(block.Block)
	if err != nil {
		log.Fatalf("Assertion Error: %s\n", err.Error())
	}

	// Step 14: Print remaining transactions to fetch
	//
	// If you want the client to automatically fetch these, consider
	// using the fetcher package.
	for _, txn := range block.OtherTransactions {
		log.Printf("Other Transaction: %+v\n", txn)
	}

	// Step 15: GetAccount
	account, rosettaErr, err := client.AccountAPI.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: primaryNetwork,
		AccountIdentifier: &types.AccountIdentifier{
			Address: "ckb1qyqxsztqvpfdyu00kt99hxgxcwr2l4z67ars5nv5pp",
		},
	})
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Step 16: Print the account
	prettyAccount, err := json.MarshalIndent(account, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Account: %s\n", string(prettyAccount))
}
