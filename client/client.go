package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"os"
	"time"
)

func main() {
	// Set up fabric connection
	wallet, err := gateway.NewFileSystemWallet("./wallet")
	if err != nil {
		fmt.Println("Failed to create wallet: ", err)
		os.Exit(1)
	}

	// Get gateway connection
	gw, err := gateway.Connect(
		gateway.WithAddress("localhost:7051"),
		gateway.WithIdentity(wallet, "admin"),
	)
	if err != nil {
		fmt.Println("Failed to connect to gateway: ", err)
		os.Exit(1)
	}

	// Create a client for interacting with the blockchain
	channelClient, err := channel.New(gw, channel.WithChannelID("assetchannel"))
	if err != nil {
		fmt.Println("Failed to create channel client: ", err)
		os.Exit(1)
	}

	// Create a new investor
	investorID := "investor1"
	balance := 10000
	investorArgs := [][]byte{[]byte(investorID), []byte(fmt.Sprintf("%d", balance))}
	_, err = channelClient.Execute(channel.Request{
		ChaincodeID: "asset-management",
		Fcn:          "CreateUser",
		Args:         investorArgs,
	})
	if err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Investor created successfully!")

	// Register a new asset
	isin := "US1234567890"
	companyName := "Tech Corp"
	assetType := "Equity"
	totalUnits := 1000
	pricePerUnit := 50
	assetArgs := [][]byte{
		[]byte(isin), []byte(companyName), []byte(assetType),
		[]byte(fmt.Sprintf("%d", totalUnits)), []byte(fmt.Sprintf("%d", pricePerUnit)),
	}
	_, err = channelClient.Execute(channel.Request{
		ChaincodeID: "asset-management",
		Fcn:          "RegisterAsset",
		Args:         assetArgs,
	})
	if err != nil {
		fmt.Printf("Failed to register asset: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Asset registered successfully!")

	// Subscribe to the asset
	unitsToSubscribe := 100
	subscribeArgs := [][]byte{
		[]byte(investorID), []byte(isin), []byte(fmt.Sprintf("%d", unitsToSubscribe)),
	}
	_, err = channelClient.Execute(channel.Request{
		ChaincodeID: "asset-management",
		Fcn:          "SubscribeAsset",
		Args:         subscribeArgs,
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to asset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Investor %s subscribed to %d units of asset %s\n", investorID, unitsToSubscribe, isin)

	// Redeem assets
	unitsToRedeem := 50
	redeemArgs := [][]byte{
		[]byte(investorID), []byte(isin), []byte(fmt.Sprintf("%d", unitsToRedeem)),
	}
	_, err = channelClient.Execute(channel.Request{
		ChaincodeID: "asset-management",
		Fcn:          "RedeemAsset",
		Args:         redeemArgs,
	})
	if err != nil {
		fmt.Printf("Failed to redeem assets: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Investor %s redeemed %d units of asset %s\n", investorID, unitsToRedeem, isin)

	// Check the investor's portfolio balance
	portfolioArgs := [][]byte{[]byte(investorID)}
	portfolioResponse, err := channelClient.Query(channel.Request{
		ChaincodeID: "asset-management",
		Fcn:          "GetPortfolio",
		Args:         portfolioArgs,
	})
	if err != nil {
		fmt.Printf("Failed to get portfolio: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Investor's Portfolio:\n", string(portfolioResponse.Payload))
}
