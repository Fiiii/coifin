package modules

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fiiii/coifin/models"
	"log"
	"math/big"
)

// GetLatestBlock - get information about the block & transactions embedded
func GetLatestBlock(client ethclient.Client) *models.Block {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	// Query the latest block
	header, _ := client.HeaderByNumber(context.Background(), nil)
	blockNumber := big.NewInt(header.Number.Int64())
	block, err := client.BlockByNumber(context.Background(), blockNumber)

	if err != nil {
		log.Fatal(err)
	}

	// Build the response to our model
	responseBlock := &models.Block{
		BlockNumber:       block.Number().Int64(),
		Timestamp:         block.Time(),
		Difficulty:        block.Difficulty().Uint64(),
		Hash:              block.Hash().String(),
		TransactionsCount: len(block.Transactions()),
		Transactions:      []models.Transaction{},
	}

	for _, tx := range block.Transactions() {
		responseBlock.Transactions = append(responseBlock.Transactions, models.Transaction{
			Hash:     tx.Hash().String(),
			Value:    tx.Value().String(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice().Uint64(),
			Nonce:    tx.Nonce(),
			To:       tx.To().String(),
		})
	}

	return responseBlock
}

// GetTxByHash by a given hash
func GetTxByHash(client ethclient.Client, hash common.Hash) *models.Transaction {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	tx, pending, err := client.TransactionByHash(context.Background(), hash)
	if err != nil {
		fmt.Println(err)
	}

	return &models.Transaction{
		Hash:     tx.Hash().String(),
		Value:    tx.Value().String(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice().Uint64(),
		To:       tx.To().String(),
		Pending:  pending,
		Nonce:    tx.Nonce(),
	}
}

// GetAddressBalance returns the given address balance
func GetAddressBalance(client ethclient.Client, address string) (string, error) {
	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return "0", err
	}

	return balance.String(), nil
}

func TransferEth(client ethclient.Client, privKey string, to string, amount int64) (string, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// Assuming you've already connected a client, the next step is to load your private key.
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return "", err
	}

	// Function requires the public address of the account we're sending from -- which we can derive from the private key.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Now we can read the nonce that we should use for the account's transaction.
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	value := big.NewInt(amount)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// We figure out who we're sending the ETH to.
	toAddress := common.HexToAddress(to)
	var data []byte

	// We create the transaction payload
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	//
	//x := types.NewTx(&types.DynamicFeeTx{
	//	Nonce: nonce,
	//	To: &toAddress,
	//	Value: value,
	//	Data: data,
	//})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	// We sign the transaction using the sender's private key
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	// Now we are finally ready to broadcast the transaction to the entire network
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	// We return the transaction hash
	return signedTx.Hash().String(), nil
}