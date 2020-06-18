package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"
	typesCKB "github.com/ququzone/ckb-sdk-go/types"
)

// BlockAPIService implements the server.BlockAPIServicer interface.
type BlockAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewBlockAPIService(network *types.NetworkIdentifier, client rpc.Client) server.BlockAPIServicer {
	return &BlockAPIService{
		network: network,
		client:  client,
	}
}

// Block implements the /block endpoint.
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	var block *typesCKB.Block
	var err error
	if request.BlockIdentifier.Hash == nil || *request.BlockIdentifier.Hash == "" {
		if *request.BlockIdentifier.Index < 0 {
			*request.BlockIdentifier.Index = 0
		}
		block, err = s.client.GetBlockByNumber(context.Background(), uint64(*request.BlockIdentifier.Index))
	} else {
		block, err = s.client.GetBlock(context.Background(), typesCKB.HexToHash(*request.BlockIdentifier.Hash))
	}
	if err != nil {
		return nil, RpcError
	}

	result := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Header.Number),
				Hash:  block.Header.Hash.String(),
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Header.Number),
				Hash:  block.Header.Hash.String(),
			},
			Timestamp:    int64(block.Header.Timestamp),
			Transactions: []*types.Transaction{},
		},
	}

	if block.Header.Number > 0 {
		result.Block.ParentBlockIdentifier = &types.BlockIdentifier{
			Index: int64(block.Header.Number) - 1,
			Hash:  block.Header.ParentHash.String(),
		}
	}

	for i, tx := range block.Transactions {
		var transaction *types.Transaction
		optIndex := int64(0)
		if i == 0 && len(tx.Outputs) > 0 {
			transaction = &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: tx.Hash.String(),
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{
							Index: optIndex,
						},
						Type:   "Reward",
						Status: "Success",
						Account: &types.AccountIdentifier{
							Address: GenerateAddress(s.network, tx.Outputs[0].Lock),
						},
						Amount: &types.Amount{
							Value:    fmt.Sprintf("%d", tx.Outputs[0].Capacity),
							Currency: CkbCurrency,
						},
					},
				},
			}
			optIndex++
		} else {
			transaction = &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: tx.Hash.String(),
				},
				Operations: []*types.Operation{},
			}
			for _, input := range tx.Inputs {
				if input.PreviousOutput.TxHash.String() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
					previousTx, err := s.client.GetTransaction(context.Background(), input.PreviousOutput.TxHash)
					if err != nil {
						return nil, RpcError
					}
					transaction.Operations = append(transaction.Operations, &types.Operation{
						OperationIdentifier: &types.OperationIdentifier{
							Index: optIndex,
						},
						Type:   "Transfer",
						Status: "Success",
						Account: &types.AccountIdentifier{
							Address: GenerateAddress(s.network, previousTx.Transaction.Outputs[input.PreviousOutput.Index].Lock),
						},
						Amount: &types.Amount{
							Value:    fmt.Sprintf("%d", previousTx.Transaction.Outputs[input.PreviousOutput.Index].Capacity),
							Currency: CkbCurrency,
						},
					})
				}
				optIndex++
			}
			for _, output := range tx.Outputs {
				transaction.Operations = append(transaction.Operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: optIndex,
					},
					Type:   "Receive",
					Status: "Success",
					Account: &types.AccountIdentifier{
						Address: GenerateAddress(s.network, output.Lock),
					},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("%d", output.Capacity),
						Currency: CkbCurrency,
					},
				})
				optIndex++
			}
		}
		result.Block.Transactions = append(result.Block.Transactions, transaction)
	}

	return result, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	tx, err := s.client.GetTransaction(context.Background(), typesCKB.HexToHash(request.TransactionIdentifier.Hash))
	if err != nil {
		return nil, RpcError
	}
	var transaction *types.Transaction
	optIndex := int64(0)
	if tx.Transaction.Inputs[0].PreviousOutput.TxHash.String() == "0x0000000000000000000000000000000000000000000000000000000000000000" && len(tx.Transaction.Outputs) > 0 {
		transaction = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: tx.Transaction.Hash.String(),
			},
			Operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{
						Index: optIndex,
					},
					Type:   "Reward",
					Status: "Success",
					Account: &types.AccountIdentifier{
						Address: GenerateAddress(s.network, tx.Transaction.Outputs[0].Lock),
					},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("%d", tx.Transaction.Outputs[0].Capacity),
						Currency: CkbCurrency,
					},
				},
			},
		}
		optIndex++
	} else {
		transaction = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: tx.Transaction.Hash.String(),
			},
			Operations: []*types.Operation{},
		}
		for _, input := range tx.Transaction.Inputs {
			previousTx, err := s.client.GetTransaction(context.Background(), input.PreviousOutput.TxHash)
			if err != nil {
				return nil, RpcError
			}
			transaction.Operations = append(transaction.Operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: optIndex,
				},
				Type:   "Transfer",
				Status: "Success",
				Account: &types.AccountIdentifier{
					Address: GenerateAddress(s.network, previousTx.Transaction.Outputs[input.PreviousOutput.Index].Lock),
				},
				Amount: &types.Amount{
					Value:    fmt.Sprintf("%d", previousTx.Transaction.Outputs[input.PreviousOutput.Index].Capacity),
					Currency: CkbCurrency,
				},
			})
			optIndex++
		}
		for _, output := range tx.Transaction.Outputs {
			transaction.Operations = append(transaction.Operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: optIndex,
				},
				Type:   "Receive",
				Status: "Success",
				Account: &types.AccountIdentifier{
					Address: GenerateAddress(s.network, output.Lock),
				},
				Amount: &types.Amount{
					Value:    fmt.Sprintf("%d", output.Capacity),
					Currency: CkbCurrency,
				},
			})
			optIndex++
		}
	}

	return &types.BlockTransactionResponse{
		Transaction: transaction,
	}, nil
}
