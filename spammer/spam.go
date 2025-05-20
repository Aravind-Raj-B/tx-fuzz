package spammer

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/crypto/slhdsa"
)

type Spam func(*Config, *slhdsa.PrivateKey, *filler.Filler) error

func SpamTransactions(config *Config, fun Spam) error {
	fmt.Printf("Spamming %v transactions per account on %v accounts with seed: 0x%x\n", config.N, len(config.keys), config.seed)

	errCh := make(chan error, len(config.keys))
	var wg sync.WaitGroup
	wg.Add(len(config.keys))
	for _, key := range config.keys {
		// Setup randomness uniquely per key
		random := make([]byte, 10000)
		config.mut.FillBytes(&random)

		var f *filler.Filler
		if len(config.corpus) != 0 {
			elem := config.corpus[rand.Int31n(int32(len(config.corpus)))]
			config.mut.MutateBytes(&elem)
			f = filler.NewFiller(elem)
		} else {
			// Use lower entropy randomness for filler
			config.mut.MutateBytes(&random)
			f = filler.NewFiller(random)
		}
		// Start a fuzzing thread
		go func(key *slhdsa.PrivateKey, filler *filler.Filler) {
			defer wg.Done()
			errCh <- fun(config, key, f)
		}(key, f)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
