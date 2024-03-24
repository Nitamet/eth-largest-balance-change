package ethereum

import (
	"context"
	"errors"
	"fmt"
	"github.com/Nitamet/eth-largest-balance-change/jsonrpc"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Transaction struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

type Block struct {
	Transactions []Transaction `json:"transactions"`
}

type Service struct {
	rpc *jsonrpc.Client
}

const (
	averageNumberOfTransactionsPerBlock = 100
	ethWeiConversionRate                = 1e18
	concurrentRequestsLimit             = 40
	rateLimitExceededCode               = 429
	rateLimitAttempts                   = 5
)

func CreateService(url, token string) (*Service, error) {
	if url == "" {
		return nil, errors.New("empty URL provided")
	}

	if token == "" {
		return nil, errors.New("empty token provided")
	}

	// Remove trailing slash if present
	url = strings.TrimSuffix(url, "/")

	return &Service{
		rpc: &jsonrpc.Client{
			Endpoint: fmt.Sprintf("%s/%s", url, token),
		},
	}, nil
}

func (s *Service) getLastBlock(ctx context.Context) (string, error) {
	var lastBlockNumber string

	err := s.rpc.Call(ctx, "eth_blockNumber", nil, &lastBlockNumber)
	if err != nil {
		return "", err
	}

	return lastBlockNumber, nil
}

func (s *Service) getBlockByNumber(ctx context.Context, blockNumber int64) (*Block, error) {
	if blockNumber < 0 {
		return nil, fmt.Errorf("invalid block number %d, expected positive value", blockNumber)
	}

	blockNumberParam := fmt.Sprintf("0x%x", blockNumber)

	var block Block

	err := s.rpc.Call(ctx, "eth_getBlockByNumber", []any{blockNumberParam, true}, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (s *Service) GetLastNBlocks(ctx context.Context, n int64) ([]Block, error) {
	if n <= 0 || n > 500 {
		return nil, fmt.Errorf("invalid number of blocks %d, allowed range is 1-500", n)
	}

	lastBlock, err := s.getLastBlock(ctx)
	if err != nil {
		return nil, err
	}

	lastBlockNumber, err := strconv.ParseInt(lastBlock, 0, 64)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Last block: %s\n\n", lastBlock)

	blocks := make([]Block, 0, n)

	var mu sync.Mutex
	var wg sync.WaitGroup
	// Use semaphore to limit the number of concurrent requests
	sem := make(chan struct{}, concurrentRequestsLimit)

	for i := lastBlockNumber; i > lastBlockNumber-n; i-- {
		wg.Add(1)
		sem <- struct{}{}

		go func(blockNumber int64) {
			defer wg.Done()

			block, err := s.tryGetBlockByNumber(ctx, blockNumber)
			<-sem

			if err != nil {
				fmt.Printf("Failed to get block %d: %s\n", blockNumber, err)

				return
			}

			mu.Lock()
			blocks = append(blocks, *block)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if len(blocks) != int(n) {
		return nil, fmt.Errorf("failed to collect all blocks, expected %d, got %d", n, len(blocks))
	}

	return blocks, nil
}

func (s *Service) tryGetBlockByNumber(ctx context.Context, blockNumber int64) (*Block, error) {
	if blockNumber < 0 {
		return nil, fmt.Errorf("invalid block number %d, expected positive value", blockNumber)
	}

	attempts := rateLimitAttempts

	for attempts > 0 {
		attempts--

		block, err := s.getBlockByNumber(ctx, blockNumber)
		if err == nil {
			return block, nil
		}

		var errInvalidResponseCode jsonrpc.InvalidResponseCodeError
		if errors.As(err, &errInvalidResponseCode) && errInvalidResponseCode.Code == rateLimitExceededCode {
			fmt.Printf("Rate limit exceeded, waiting 1 second and retrying...\n\n")
			time.Sleep(1 * time.Second)
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("failed to collect transactions for block %d, no more attempts", blockNumber)
}

func (s *Service) GetLargestBalanceChange(blocks []Block) (string, *big.Int, error) {
	if len(blocks) == 0 {
		return "", nil, errors.New("no blocks provided")
	}

	addressesBalanceChanges := make(map[string]*big.Int, len(blocks)*averageNumberOfTransactionsPerBlock)

	for _, block := range blocks {
		err := s.calculateTransactions(block.Transactions, addressesBalanceChanges)

		if err != nil {
			return "", nil, err
		}
	}

	largestChange := big.NewInt(0)
	largestAddress := ""

	for address, balance := range addressesBalanceChanges {
		balanceAbs := big.NewInt(0).Abs(balance)
		largestChangeAbs := big.NewInt(0).Abs(largestChange)

		if balanceAbs.Cmp(largestChangeAbs) == 1 {
			largestChange = balance
			largestAddress = address
		}
	}

	return largestAddress, largestChange, nil
}

func (s *Service) calculateTransactions(transactions []Transaction, addressesBalanceChanges map[string]*big.Int) error {
	if len(transactions) == 0 {
		return errors.New("no transactions provided")
	}

	if addressesBalanceChanges == nil {
		return errors.New("nil map provided")
	}

	for _, transaction := range transactions {
		transactionValue, ok := big.NewInt(0).SetString(transaction.Value, 0)
		if !ok {
			return fmt.Errorf("failed to parse transaction value: %s", transaction.Value)
		}

		addressFromBalance, hasAddressFrom := addressesBalanceChanges[transaction.From]
		if !hasAddressFrom {
			zeroValue := big.NewInt(0)
			addressesBalanceChanges[transaction.From] = zeroValue.Sub(zeroValue, transactionValue)
		} else {
			addressFromBalance.Sub(addressFromBalance, transactionValue)
		}

		addressToBalance, hasAddressTo := addressesBalanceChanges[transaction.To]
		if !hasAddressTo {
			zeroValue := big.NewInt(0)
			addressesBalanceChanges[transaction.To] = zeroValue.Add(zeroValue, transactionValue)
		} else {
			addressToBalance.Add(addressToBalance, transactionValue)
		}
	}

	return nil
}

func WeiToEth(wei *big.Int) (*big.Float, error) {
	if wei == nil {
		return nil, errors.New("nil wei provided")
	}

	eth := new(big.Float).SetInt(wei)

	return eth.Quo(eth, new(big.Float).SetInt(big.NewInt(ethWeiConversionRate))), nil
}
