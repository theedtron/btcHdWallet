package main

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/tyler-smith/go-bip39"
)

func main() {
	// Generate a new mnemonic seed
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		fmt.Println("Error generating entropy:", err)
		return
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		fmt.Println("Error generating mnemonic:", err)
		return
	}

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Println("Error generating master key:", err)
	}

	fmt.Println("Master key:", masterKey)

	// Derive the first child key
	childKey, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		fmt.Println("Error deriving child key:", err)
		return
	}

	// Convert child key to Bitcoin address
	childPubKey, err := childKey.ECPubKey()
	if err != nil {
		fmt.Println("Error generating address:", err)
		return
	}

	pubKeyHash := btcutil.Hash160(childPubKey.SerializeCompressed())
	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Println("Error generating address:", err)
		return
	}

	fmt.Println("Segwit Address:", address.EncodeAddress())
	os.Exit(0)

	// Connect to Bitcoin node
	rpcClient, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         "localhost:8334",
		User:         "username",
		Pass:         "xxxxxxxxx",
	}, nil)
	if err != nil {
		fmt.Println("Error connecting to node:", err)
		return
	}

	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add input to the transaction
	prevOutHashStr := "aabbccdd112233445566778899aabbccdd112233445566778899aabbccdd1122"
	prevOutHash, err := chainhash.NewHashFromStr(prevOutHashStr)
	if err != nil {
		fmt.Println("Error parsing previous output hash:", err)
	}

	prevOut := wire.NewOutPoint(prevOutHash, 0)
	txIn := wire.NewTxIn(prevOut, nil, nil)
	tx.AddTxIn(txIn)

	// Add output to the transaction
	outputAddr, err := btcutil.DecodeAddress("tb1q9j37z2ssgvpv3uqz6d0hcg3a5nc8duq0jw6c0l", &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Println("Error decoding output address:", err)
	}

	outputScript, err := txscript.PayToAddrScript(outputAddr)
	if err != nil {
		fmt.Println("Error creating output script:", err)
	}

	txOut := wire.NewTxOut(100000, outputScript)
	tx.AddTxOut(txOut)

	// Sign the transaction
	hashType := txscript.SigHashAll
	privateKey, err := childKey.ECPrivKey() // Here we get the private key from the child key
	if err != nil {
		fmt.Println("Error obtaining private key:", err)
	}
	
	script := outputScript
	sigScript, err := txscript.SignatureScript(tx, 0, script, hashType, privateKey, true)
	if err != nil {
		fmt.Println("Error signing transaction:", err)
	}

	txIn.SignatureScript = sigScript

	fmt.Printf("Signed Transaction: %v\n", tx)

	// Send the transaction
	txHash, err := rpcClient.SendRawTransaction(tx, false)
	if err != nil {
		fmt.Println("Error sending transaction:", err)
		return
	}

	fmt.Println("Transaction sent! TxHash:", txHash.String())
}
