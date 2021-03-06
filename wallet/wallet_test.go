package wallet

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	wi "github.com/OpenBazaar/wallet-interface"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

var validRopstenURL = fmt.Sprintf("https://ropsten.infura.io/%s", validInfuraKey)

var validRopstenWallet *EthereumWallet
var destWallet *EthereumWallet

var script EthRedeemScript

func setupRopstenWallet() {
	validRopstenWallet = NewEthereumWallet(validRopstenURL, validKeyFile, validPassword)
}

func setupDestWallet() {
	destWallet = NewEthereumWallet(validRopstenURL,
		"../test/UTC--2018-06-16T20-09-33.726552102Z--cecb952de5b23950b15bfd49302d1bdd25f9ee67", validPassword)
}

func setupEthRedeemScript(timeout time.Duration, threshold int) {

	script.TxnID = common.StringToAddress(xid.New().String() + xid.New().String())
	script.Timeout = uint32(timeout.Hours())
	script.Threshold = uint8(threshold)
	script.Buyer = common.HexToAddress(validSourceAddress)
	script.Seller = common.HexToAddress(validDestinationAddress)
}

func TestNewWalletWithValidValues(t *testing.T) {
	wallet := NewEthereumWallet(validRopstenURL, validKeyFile, validPassword)
	if wallet == nil {
		t.Errorf("valid credentials should return a wallet")
	}
	if wallet.address.String() != validSourceAddress {
		t.Errorf("valid credentials should return a wallet with proper initialization")
	}
}

func TestNewWalletWithInValidValues(t *testing.T) {
	t.SkipNow()
	wallet := NewEthereumWallet(validRopstenURL, validKeyFile, invalidPassword)
	if wallet != nil {
		t.Errorf("invalid credentials should return a wallet")
	}
}

func TestWalletGetBalance(t *testing.T) {
	setupRopstenWallet()

	if _, err := validRopstenWallet.GetBalance(); err != nil {
		t.Errorf("valid wallet should return balance")
	}
}

func TestWalletGetUnconfirmedBalance(t *testing.T) {
	setupRopstenWallet()

	if _, err := validRopstenWallet.GetUnconfirmedBalance(); err != nil {
		t.Errorf("valid wallet should return unconfirmed balance")
	}
}

func TestWalletTransfer(t *testing.T) {
	//t.SkipNow()
	setupRopstenWallet()
	setupDestWallet()

	value := big.NewInt(200000)

	sbal1 := big.NewInt(0)
	dbal1 := big.NewInt(0)

	cbal1, _ := validRopstenWallet.GetBalance()
	ucbal1, _ := validRopstenWallet.GetUnconfirmedBalance()

	cbal2, _ := destWallet.GetBalance()
	ucbal2, _ := destWallet.GetUnconfirmedBalance()

	sbal1.Add(cbal1, ucbal1)
	dbal1.Add(cbal2, ucbal2)

	_, err := validRopstenWallet.Transfer(validDestinationAddress, value)

	if err != nil {
		t.Errorf("valid wallet should allow transfer")
	}

	//_, err = chainhash.NewHashFromStr(hash.String())

	//if err != nil {
	//	t.Errorf("wallet should return a valid transaction")
	//}

	//txn, err := validRopstenWallet.GetTransaction(*chash)

	//if err != nil {
	//	t.Errorf("wallet should return a valid transaction")
	//}

	//if txn.Value != value.Int64() {
	//	t.Errorf("wallet is not forming the correct txn")
	//}

	sbal2 := big.NewInt(0)
	dbal2 := big.NewInt(0)

	cbal1, _ = validRopstenWallet.GetBalance()
	ucbal1, _ = validRopstenWallet.GetUnconfirmedBalance()

	cbal2, _ = destWallet.GetBalance()
	ucbal2, _ = destWallet.GetUnconfirmedBalance()

	sbal2.Add(cbal1, ucbal1)
	dbal2.Add(cbal2, ucbal2)

	val := big.NewInt(0)

	val.Sub(dbal2, dbal1)

	if val.Cmp(value) != 0 {
		t.Errorf("client should have transferred balance")
	}

}

func TestWalletCurrencyCode(t *testing.T) {
	setupRopstenWallet()

	if validRopstenWallet.CurrencyCode() != "ETH" {
		t.Errorf("wallet should return proper currency code")
	}
}

func TestWalletIsDust(t *testing.T) {
	setupRopstenWallet()

	if validRopstenWallet.IsDust(int64(10000 + 10000)) {
		t.Errorf("wallet should not indicate wrong dust")
	}

	if !validRopstenWallet.IsDust(int64(10000 - 100)) {
		t.Errorf("wallet should not indicate wrong dust")
	}
}

func TestWalletCurrentAddress(t *testing.T) {
	setupRopstenWallet()

	addr := validRopstenWallet.CurrentAddress(wi.EXTERNAL)

	if addr.String() != validSourceAddress {
		t.Errorf("wallet should return correct current address")
	}
}

func TestWalletNewAddress(t *testing.T) {
	setupRopstenWallet()

	addr := validRopstenWallet.NewAddress(wi.EXTERNAL)

	if addr.String() != validSourceAddress {
		t.Errorf("wallet should return correct new address")
	}
}

func TestWalletContractAddTransaction(t *testing.T) {
	setupRopstenWallet()
	d, _ := time.ParseDuration("1h")
	setupEthRedeemScript(d, 1)

	redeemScript, err := SerializeEthScript(script)
	if err != nil {
		t.Error("error serializing redeem script")
	}

	hash := sha3.NewKeccak256()
	hash.Write(redeemScript)
	hashStr := hexutil.Encode(hash.Sum(nil)[:])
	shash1 := crypto.Keccak256(redeemScript)
	shash1Str := hexutil.Encode(shash1)
	fmt.Println("hashStr : ", hashStr)
	fmt.Println("shash1Str : ", shash1Str)
	addr := common.HexToAddress(hashStr)
	var shash [32]byte
	copy(shash[:], addr.Bytes())

	var s1 [32]byte
	var s2 [32]byte

	copy(s1[:], hash.Sum(nil)[:])
	copy(s2[:], shash1)

	fmt.Println("s1 and s2 are equal? : ", bytes.Equal(s1[:], s2[:]))

	fromAddress := validRopstenWallet.account.Address()
	nonce, err := validRopstenWallet.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := validRopstenWallet.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(validRopstenWallet.account.key.PrivateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(66778899)   // in wei
	auth.GasLimit = big.NewInt(4000000) // in units
	auth.GasPrice = gasPrice

	fmt.Println("buyer : ", script.Buyer)
	fmt.Println("seller : ", script.Seller)
	fmt.Println("moderator : ", script.Moderator)
	fmt.Println("threshold : ", script.Threshold)
	fmt.Println("timeout : ", script.Timeout)
	fmt.Println("scrptHash : ", shash)

	/*
		var tx *types.Transaction

		if script.Threshold == 1 {
			tx, err = validRopstenWallet.ppsct.AddTransaction(auth, script.Buyer, script.Seller,
				[]common.Address{}, script.Threshold, script.Timeout, shash)
		} else {
			tx, err = validRopstenWallet.ppsct.AddTransaction(auth, script.Buyer, script.Seller,
				[]common.Address{script.Moderator}, script.Threshold, script.Timeout, shash)
		}

		fmt.Println(tx)
		fmt.Println(err)
	*/
}
