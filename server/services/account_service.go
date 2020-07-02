package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/indexer"
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

// NewAccountAPIService creates a new instance of a AccountAPIService.
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

	header, err := s.client.GetTip(context.Background())
	if err != nil {
		return nil, RpcError
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

	if total > 0 {
		if cells.Objects[len(cells.Objects)-1].BlockNumber > header.BlockNumber {
			header, err := s.client.GetHeaderByNumber(context.Background(), cells.Objects[len(cells.Objects)-1].BlockNumber)
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
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(header.BlockNumber),
			Hash:  header.BlockHash.String(),
		},
		Balances: []*types.Amount{
			{
				Value:    fmt.Sprintf("%d", total),
				Currency: CkbCurrency,
			},
		},
	}, nil
}
