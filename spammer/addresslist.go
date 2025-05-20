package spammer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/slhdsa"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	staticKeys = []string{
		"0x567ecd8db7cde838f4ff8390e7343fadae8a67507d979088cab597dbcc0b6e1127bfd57a98a691eca264a2945193e0470bc27ef12373f7b978f4e565d8fc10fb8049c8a674e3b4cb13418c6c3fe6fc73c98a0aa74a573cb5ea7b5505eb227acb70bdb184ac9658cb8a4d63ea4536c6de62ec66ee2af9314462d748a1306ce1cd",
		"0x73ed26acae7f5d3cb954086c36a8d2256ca08224bf7924229af545eb26c46d09c994a8cbfe359dca96e65f340e3ff08a2ab3d755f3a4f48517182b6b4835399ff84b554dfd6036990802992119fa1c941552108263721d43fef7b642e40a1ee1fc8a1329a4268b964e98a31e1da8368945cec0f550fd0a40bc7c3950de3a6049",
		"0x31a678304a70d4707c1d56075c9a202193eb2934e8188a691f6fc81d395d71a7e793bed7d700a484a3dd14b319f85920831f61f1b3ebd7c958227dd806ed1421a0291696b758fe269b08a5e9adc0cef9af1dee681f8d3b8eea878fe193ae7db63a471406aec742f61a34055383d194a687f7b1041de7341a6d8c4fdcb2613f00",
		"0x48825ac653d0c2e34017feb6d1c6917a61f6d1e99d6e22c7f02e7ab0cfd33747cacf44c95653972ddf51f777e536de4320a13930d10abbc71dd74ea4e46fa112ee2bdd197be2df23c679f04322ce9e094643059010ca3a94f5dfcfcd149a3f526a62bcd1b8413b998176ad1f1ed122c1da9f2dac09b54820f212130b06020d4e",
		"0x693cc9e1637a3fcd56c7dadf3bb5bbd38b811deeaad600d2668acb43f2d1b5621f21041d99af14205c6456a3b9f8ab215afe289471a88f502c5dbf4706b5870fcbd1e2d2bcfbcd20ee8266619f863d266c4938a51a656bb38576bedda6641322add9c6e65598233bec5f0e2a4d9716f98e54e1a3365fccb03bd96d8012bc9e34",
		"0x6c1c823defb1796664359bab6644180d831817ac3f45487ad6b9fdd9b37f779eb5b0f741abf8e3712780594fdbdb62017d40cfd3ce7d668e5c1b76eaa03f0e9a5f41b4ea2b16d50de0ebf1261d10b3a6076094efc4f14e3916d8b438c3d332fff1a58143bbb5169cc1212881174be971565c3ac417f470f40c49c1a019b4c4b2",
		"0x7f851e591f398c4d10ebda4ce213c3b4497462fa519f01a6b81633fece44960213ac3e2233abb2d13b1e46ca8184c4e2a2ab0c110a6703e04e3777b72094b6c17b295d219c8bb2e6cba4d2ea88ee33e837d614e3bc4ed4fc8694ccee672102033de449c48b8ecb4c3c4d49400ff8233a59263d1f9a6a647d6ec5eecc261c40cd",
		"0x4f6d9a56835e4fbaa9af663ec3704fcb0ebe47edf5db1ceec5121b64c9baa8c1cce77ebfd32b619256d8dd7e7caf7011acbbf9b2637520560d90f030e043b748263792063eea7d8e64f3cf0da8ebaea3041624622d6c9ecbce5a8159b3483b66c51946658f3f3d9f09d20fc4297a6da994ed83a237e5259ece9f346424745418",
		"0xc73938579570f1ed36377b70a0c8976750354719c950472dbf385fadda8565a1fa0a78ac71000f05330c1bfffb424df3715c9cfd1d92294a7388cae4a12afc9117c0e199b4698b7206e58ecf8d974e1f34824611ecdcc95280c238f57e28fd6f8bcba3a982875c99d77dfb9f4351279ef3c8b0bf1e191841edcbbc4cb1355097",
	}
)

func CreateAddresses(N int) ([]string, []string) {
	keys := make([]string, 0, N)
	addrs := make([]string, 0, N)

	k := CreateAddressesRaw(N)
	for _, sk := range k {
		addr := crypto.PubkeyToAddress(sk.PublicKey)
		skHex := "0x" + common.Bytes2Hex(crypto.FromSLHDSA(sk))
		// Sanity check marshalling
		skTest, err := crypto.ToSLHDSA(crypto.FromSLHDSA(sk))
		if err != nil {
			panic(err)
		}
		_ = skTest
		keys = append(keys, skHex)
		addrs = append(addrs, addr.Hex())
	}
	return keys, addrs
}

func CreateAddressesRaw(N int) []*slhdsa.PrivateKey {
	keys := make([]*slhdsa.PrivateKey, 0, N)

	for i := 0; i < N; i++ {
		// WARNING= USES UNSECURE RANDOMNESS
		sk, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		keys = append(keys, sk)
	}
	return keys
}

func Airdrop(config *Config, value *big.Int) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(config.faucet.PublicKey)
	fmt.Printf("Airdrop faucet is at %x\n", sender)
	var tx *types.Transaction
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		fmt.Printf("error getting chain ID; could not airdrop: %v\n", err)
		return err
	}
	for _, addr := range config.keys {
		nonce, err := backend.PendingNonceAt(context.Background(), sender)
		if err != nil {
			fmt.Printf("error getting pending nonce; could not airdrop: %v\n", err)
			return err
		}
		to := crypto.PubkeyToAddress(addr.PublicKey)
		gp, _ := backend.SuggestGasPrice(context.Background())
		gas, err := backend.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     crypto.PubkeyToAddress(config.faucet.PublicKey),
			To:       &to,
			Gas:      30_000_000,
			GasPrice: gp,
			Value:    value,
		})
		if err != nil {
			fmt.Printf("error estimating gas: %v\n", err)
			fmt.Printf("estimating: from %v, to %v, gas %v, gasprice %v value %v", crypto.PubkeyToAddress(config.faucet.PublicKey), &to, 30_000_000, gp, value)
			return err
		}
		tx2 := types.NewTransaction(nonce, to, value, gas, gp, nil)
		fmt.Printf("Using chain ID: %v for signing\n", chainid)
		signedTx, _ := types.SignTx(tx2, types.LatestSignerForChainID(chainid), config.faucet)
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			fmt.Printf("error sending transaction; could not airdrop: %v\n", err)
			return err
		}
		tx = signedTx
		time.Sleep(10 * time.Millisecond)
	}
	// Wait for the last transaction to be mined
	if _, err := bind.WaitMined(context.Background(), backend, tx); err != nil {
		return err
	}
	return nil
}
