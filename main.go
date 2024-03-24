package main

import (
	"context"
	"fmt"
	"github.com/Nitamet/eth-largest-balance-change/ethereum"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicf("Error loading .env file: %v", err)
	}

	endpointUrl := os.Getenv("ENDPOINT_URL")
	token := os.Getenv("API_TOKEN")

	ethereumService := ethereum.CreateService(endpointUrl, token)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var blocksToFetch int64 = 100
	blocks, err := ethereumService.GetLastNBlocks(ctx, blocksToFetch)
	if err != nil {
		panic(err)
	}

	address, value, err := ethereumService.GetLargestBalanceChange(blocks)
	if err != nil {
		panic(err)
	}

	zero := big.NewInt(0)

	fmt.Printf("The address that sent/received the most in the last %d blocks:\n", blocksToFetch)
	if value.Cmp(zero) == -1 {
		fmt.Printf("Address %s sent the most\n", address)
	} else {
		fmt.Printf("Address %s received the most\n", address)
	}

	fmt.Println(value.String(), "WEI")

	eth, err := ethereum.WeiToEth(value)
	if err != nil {
		panic(err)
	}

	fmt.Println(eth.Text('f', 6), "ETH")
}
