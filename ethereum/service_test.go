package ethereum_test

import (
	"github.com/Nitamet/eth-largest-balance-change/ethereum"
	"math/big"
	"testing"
)

// blockTransactions is a helper struct to create ethereum.Block from a slice of transactions.
// 1st index is the sender address.
// 2nd index is the receiver address.
// 3rd index is the value of the transaction.
type blockTransactions struct {
	transactions [][3]string
}

func createBlockFromBlockTransactions(data blockTransactions) ethereum.Block {
	transactions := make([]ethereum.Transaction, 0, len(data.transactions))
	for _, transaction := range data.transactions {
		transactions = append(transactions, ethereum.Transaction{
			From:  transaction[0],
			To:    transaction[1],
			Value: transaction[2],
		})
	}

	return ethereum.Block{Transactions: transactions}
}

func TestService_GetLargestBalanceChange(t *testing.T) {
	t.Run("should return the address with the largest balance change", func(t *testing.T) {
		transactionsData := []blockTransactions{
			{
				transactions: [][3]string{
					{"0x1", "0x2", "1535333445"},
					{"0x1", "0x2", "4545666742"},
					{"0x2", "0x3", "97979882342"},
				},
			},
			{
				transactions: [][3]string{
					{"0x3", "0x5", "153445534656"},
					{"0x7", "0x5", "2341545"},
					{"0x3", "0x1", "545321498"},
				},
			},
			{
				transactions: [][3]string{
					{"0x5", "0x1", "13215523"},
					{"0x4", "0x7", "5366678"},
				},
			},
		}

		blocks := make([]ethereum.Block, 0, 3)
		for _, data := range transactionsData {
			blocks = append(blocks, createBlockFromBlockTransactions(data))
		}

		service := ethereum.CreateService("", "")

		expectedAddress := "0x5"
		expectedValue := big.NewInt(153434660678)

		gotAddress, gotValue, err := service.GetLargestBalanceChange(blocks)
		if err != nil {
			t.Fatal(err)
		}

		if gotAddress != expectedAddress {
			t.Fatalf("expected %s, got %s", expectedAddress, gotAddress)
		}

		if expectedValue.Cmp(gotValue) != 0 {
			t.Fatalf("expected %d, got %d", expectedValue, gotValue)
		}
	})
}
