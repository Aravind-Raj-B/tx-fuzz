package spammer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/slhdsa"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const TX_TIMEOUT = 5 * time.Minute

func SendBasicTransactions(config *Config, key *slhdsa.PrivateKey, f *filler.Filler) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(key.PublicKey)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainID, using default")
		chainID = big.NewInt(4062024)
	}

	var lastTx *types.Transaction
	for i := uint64(0); i < config.N; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			return err
		}
		tx, err := txfuzz.RandomValidTx(config.backend, f, sender, nonce, nil, nil, config.accessList)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			return err
		}

		if tx.To() == nil {
			log.Warn("Skipping transaction with nil 'To' address")
			continue
		}

		// fmt.Printf("\nTransaction Parameters:\n")
		// fmt.Printf("Nonce: %d\n", tx.Nonce())
		// if tx.To() == nil {
		// 	fmt.Printf("To: %s\n", "Nill")
		// } else {
		// 	fmt.Printf("To: %s\n", tx.To().Hex())
		// }
		// //fmt.Printf("To: %s\n", tx.To().Hex())
		// fmt.Printf("Value: %s\n", tx.Value().String())
		// fmt.Printf("Gas: %d\n", tx.Gas())
		// fmt.Printf("GasPrice: %s\n", tx.GasPrice().String())
		// fmt.Printf("GasFeeCap: %s\n", tx.GasFeeCap().String())
		// fmt.Printf("GasTipCap: %s\n", tx.GasTipCap().String())
		// fmt.Printf("Data: %x\n", tx.Data())
		// fmt.Printf("ChainID: %s\n", chainID.String())
		// fmt.Printf("AccessList: %s\n", tx.AccessList())

		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), key)
		if err != nil {
			return err
		}

		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			log.Warn("Could not submit transaction: %v", err)
			return err
		}

		lastTx = signedTx
		time.Sleep(10 * time.Millisecond)
	}
	if lastTx != nil {
		ctx, cancel := context.WithTimeout(context.Background(), TX_TIMEOUT)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
			fmt.Printf("Waiting for transactions to be mined failed: %v\n", err.Error())
		}
	}
	return nil
}
