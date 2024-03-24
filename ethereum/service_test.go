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

		service, err := ethereum.CreateService("url", "token")
		if err != nil {
			t.Fatal(err)
		}

		// Address 0x1 has 0 - 1535333445 - 4545666742 + 545321498 + 13215523 =  -5 522 463 166
		// Address 0x2 has 1535333445 + 4545666742 - 97979882342 = -91 898 882 155
		// Address 0x3 has 97979882342 - 153445534656 - 545321498 = -56 010 973 812
		// Address 0x4 has 0 - 5366678 = -5 366 678
		// Address 0x5 has 153445534656 + 2341545 - 13215523 = 153 434 660 678
		// Address 0x7 has 0 - 2341545 + 5366678 = 3 025 133
		//
		// Therefore, address 0x5 has the largest balance change with 153 434 660 678
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
