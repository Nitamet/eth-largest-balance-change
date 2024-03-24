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

	ethereumService, err := ethereum.CreateService(endpointUrl, token)
	if err != nil {
		panic(err)
	}

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

	fmt.Printf(
		"The address which balance has changed the most in the last %d blocks is %s\n",
		blocksToFetch,
		address,
	)

	if value.Cmp(zero) == -1 {
		fmt.Println("It has sent the most ETH:")
	} else {
		fmt.Println("It has received the most ETH:")
	}

	fmt.Println(value.String(), "WEI")

	eth, err := ethereum.WeiToEth(value)
	if err != nil {
		panic(err)
	}

	fmt.Println(eth.Text('f', 6), "ETH")
}
