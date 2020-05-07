package services

import (
	"context"
	"fmt"
	"github.com/ququzone/ckb-rich-sdk-go/indexer"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"
	"github.com/ququzone/ckb-sdk-go/address"
)

const (
	pageSize = 10000000
)

// AccountAPIService implements the server.AccountAPIServicer interface.
type AccountAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewAccountAPIService(network *types.NetworkIdentifier, client rpc.Client) server.AccountAPIServicer {
	return &AccountAPIService{
		network: network,
		client:  client,
	}
}

// AccountBalance implements the /account/balance endpoint.
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
	addr, err := address.Parse(request.AccountIdentifier.Address)
	if err != nil {
		return nil, AddressError
	}

	var total uint64
	cells, err := s.client.GetCells(context.Background(), &indexer.SearchKey{
		Script:     addr.Script,
		ScriptType: indexer.ScriptTypeLock,
	}, indexer.SearchOrderAsc, pageSize, "")
	if err != nil {
		return nil, RpcError
	}
	for _, cell := range cells.Objects {
		total += cell.Output.Capacity
	}
	for ; len(cells.Objects) == pageSize; {
		cells, err = s.client.GetCells(context.Background(), &indexer.SearchKey{
			Script:     addr.Script,
			ScriptType: indexer.ScriptTypeLock,
		}, indexer.SearchOrderAsc, pageSize, cells.LastCursor)
		if err != nil {
			return nil, RpcError
		}
		for _, cell := range cells.Objects {
			total += cell.Output.Capacity
		}
	}

	header, err := s.client.GetTipHeader(context.Background())
	if err != nil {
		return nil, RpcError
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(header.Number),
			Hash:  header.Hash.String(),
		},
		Balances: []*types.Amount{
			{
				Value:    fmt.Sprintf("%d", total),
				Currency: CkbCurrency,
			},
		},
	}, nil
}
