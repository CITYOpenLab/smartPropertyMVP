import(
  "fmt"
  "log"
  "net/http"

  "github.com/stellar/go/build"
  "github.com/stellar/go/clients/horizon"
  "github.com/stellar/go/keypair"
  "github.com/stellar/go/network"
)

type Account struct {
	Seed string
	PublicKey string
}

func (a *Account) New() (error) {
	pair, err := keypair.Random()
	// so key value pairs over here are ed25519 key pairs instead of bitcoin style key pairs
	// they also seem to sue al lcaps, which I don't know why
	if err != nil {
		return err
	}
	log.Println("MY SEED IS: ", pair.Seed())
	a.Seed = pair.Seed()
	a.PublicKey = pair.Address()
	return nil
}

func (a *Account) GetCoins() (error) {
	// get some coins from the stellar robot for testing
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + a.PublicKey)
	if err != nil || resp == nil {
		log.Println("ERRORED OUT while calling friendbot, no coins for us")
		return err
	}
	return nil
}

func (a *Account) Balance() (error) {

	account, err := horizon.DefaultTestNetClient.LoadAccount(a.PublicKey)
	if err != nil {
		return nil
	}

	for _, balance := range account.Balances {
		log.Println("BALANCE for account: ", a.PublicKey, " is: ", balance)
	}

	return nil
}

func (a *Account) SendCoins(destination string, amount string) (int32, string, error) {

	if _, err := horizon.DefaultTestNetClient.LoadAccount(destination); err != nil {
		// if destination doesn't exist, do nothing
		// returning -11 since -1 maybe returned for unconfirmed tx or something like that
		return -11, "", err
	}

	passphrase := network.TestNetworkPassphrase

	tx, err := build.Transaction(
		build.Network{passphrase},
		build.SourceAccount{a.Seed},
		build.AutoSequence{horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{destination},
			build.NativeAmount{amount},
		),
	)

	if err != nil {
		return -11, "", err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(a.Seed)
	if err != nil {
		return -11, "", err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		return -11, "", err
	}
	// And finally, send it off to Stellar!
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return -11, "", err
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)
	return resp.Ledger, resp.Hash, nil
}