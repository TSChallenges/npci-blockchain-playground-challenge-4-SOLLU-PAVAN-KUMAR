package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

type Asset struct {
	ISIN           string `json:"isin"`
	CompanyName    string `json:"company_name"`
	AssetType      string `json:"asset_type"`
	TotalUnits     int    `json:"total_units"`
	PricePerUnit   int    `json:"price_per_unit"`
	AvailableUnits int    `json:"available_units"`
}

type Investor struct {
	InvestorID string            `json:"investor_id"`
	Balance    int               `json:"balance"`
	Subscribed map[string]int    `json:"subscribed"` // Key: ISIN, Value: Number of units
}

type AssetManagementContract struct {
	contractapi.Contract
}

// CreateUser function adds an investor to the ledger
func (s *AssetManagementContract) CreateUser(ctx contractapi.TransactionContextInterface, investorID string, balance int) error {
	investor := Investor{
		InvestorID: investorID,
		Balance:    balance,
		Subscribed: make(map[string]int),
	}

	investorJSON, err := json.Marshal(investor)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(investorID, investorJSON)
}

// RegisterAsset function registers a new asset
func (s *AssetManagementContract) RegisterAsset(ctx contractapi.TransactionContextInterface, isin string, companyName string, assetType string, totalUnits int, pricePerUnit int) error {
	asset := Asset{
		ISIN:           isin,
		CompanyName:    companyName,
		AssetType:      assetType,
		TotalUnits:     totalUnits,
		PricePerUnit:   pricePerUnit,
		AvailableUnits: totalUnits,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(isin, assetJSON)
}

// SubscribeAsset function allows an investor to subscribe to asset units
func (s *AssetManagementContract) SubscribeAsset(ctx contractapi.TransactionContextInterface, investorID string, isin string, units int) error {
	// Get the role of the current user
	userRole, err := ctx.GetClientIdentity().GetAttributeValue("role")
	if err != nil || userRole != "Investor" {
		return fmt.Errorf("only investors can subscribe to assets")
	}

	// Fetch the investor from the ledger
	investorJSON, err := ctx.GetStub().GetState(investorID)
	if err != nil {
		return fmt.Errorf("failed to read investor: %v", err)
	}
	if investorJSON == nil {
		return fmt.Errorf("investor not found")
	}

	var investor Investor
	err = json.Unmarshal(investorJSON, &investor)
	if err != nil {
		return err
	}

	// Fetch the asset from the ledger
	assetJSON, err := ctx.GetStub().GetState(isin)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if assetJSON == nil {
		return fmt.Errorf("asset not found")
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	// Check if the investor has sufficient balance to subscribe
	totalCost := units * asset.PricePerUnit
	if investor.Balance < totalCost {
		return fmt.Errorf("insufficient balance to subscribe")
	}

	// Check if enough units are available
	if units > asset.AvailableUnits {
		return fmt.Errorf("not enough available units")
	}

	// Update the investor's subscribed units and balance
	investor.Subscribed[isin] += units
	investor.Balance -= totalCost

	// Update the asset's available units
	asset.AvailableUnits -= units

	// Save the updated investor and asset
	investorJSON, err = json.Marshal(investor)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(investorID, investorJSON)
	if err != nil {
		return err
	}

	assetJSON, err = json.Marshal(asset)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(isin, assetJSON)
	if err != nil {
		return err
	}

	// Emit a subscription event
	err = ctx.GetStub().SetEvent("SubscriptionEvent", []byte(fmt.Sprintf("Investor %s subscribed to %d units of asset %s", investorID, units, isin)))
	if err != nil {
		return err
	}

	return nil
}

// RedeemAsset function allows an investor to redeem asset units
func (s *AssetManagementContract) RedeemAsset(ctx contractapi.TransactionContextInterface, investorID string, isin string, units int) error {
	
	// Get the role of the current user
	userRole, err := ctx.GetClientIdentity().GetAttributeValue("role")
	if err != nil || userRole != "Investor" {
		return fmt.Errorf("only investors can Redeem the assets")
	}
	
	
	// Fetch the investor from the ledger
	investorJSON, err := ctx.GetStub().GetState(investorID)
	if err != nil {
		return fmt.Errorf("failed to read investor: %v", err)
	}
	if investorJSON == nil {
		return fmt.Errorf("investor not found")
	}

	var investor Investor
	err = json.Unmarshal(investorJSON, &investor)
	if err != nil {
		return err
	}

	// Fetch the asset from the ledger
	assetJSON, err := ctx.GetStub().GetState(isin)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if assetJSON == nil {
		return fmt.Errorf("asset not found")
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	// Check if the investor has enough subscribed units to redeem
	if investor.Subscribed[isin] < units {
		return fmt.Errorf("insufficient units to redeem")
	}

	// Update the investor's subscribed units and balance
	investor.Subscribed[isin] -= units
	investor.Balance += units * asset.PricePerUnit

	// Update the asset's available units
	asset.AvailableUnits += units

	// Save the updated investor and asset
	investorJSON, err = json.Marshal(investor)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(investorID, investorJSON)
	if err != nil {
		return err
	}

	assetJSON, err = json.Marshal(asset)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(isin, assetJSON)
	if err != nil {
		return err
	}

	// Emit a redemption event
	err = ctx.GetStub().SetEvent("RedemptionEvent", []byte(fmt.Sprintf("Investor %s redeemed %d units of asset %s", investorID, units, isin)))
	if err != nil {
		return err
	}

	return nil
}

// GetPortfolio function retrieves an investor's portfolio
func (s *AssetManagementContract) GetPortfolio(ctx contractapi.TransactionContextInterface, investorID string) (string, error) {
	// Fetch the investor from the ledger
	investorJSON, err := ctx.GetStub().GetState(investorID)
	if err != nil {
		return "", fmt.Errorf("failed to read investor: %v", err)
	}
	if investorJSON == nil {
		return "", fmt.Errorf("investor not found")
	}

	var investor Investor
	err = json.Unmarshal(investorJSON, &investor)
	if err != nil {
		return "", err
	}

	portfolio := fmt.Sprintf("Investor ID: %s\nBalance: %d\nSubscribed Assets:\n", investor.InvestorID, investor.Balance)
	for isin, units := range investor.Subscribed {
		portfolio += fmt.Sprintf("- Asset: %s, Units: %d\n", isin, units)
	}

	return portfolio, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AssetManagementContract{})
	if err != nil {
		fmt.Printf("Error creating asset management chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting asset management chaincode: %v", err)
	}
}
