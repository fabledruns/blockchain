package wallet

import (
	"bytes"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"os"
)

const walletFile = "./tmp/wallets.dat"

type Wallets struct {
	Wallets map[string]*Wallet
}

type walletDTO struct {
	PrivateKey []byte
	PublicKey  []byte
}

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{
		Wallets: make(map[string]*Wallet),
	}

	err := wallets.LoadFile()

	return &wallets, err
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())

	ws.Wallets[address] = wallet

	return address
}

func (ws *Wallets) SaveFile() {
	var content bytes.Buffer

	dto := make(map[string]walletDTO)

	for address, wallet := range ws.Wallets {
		der, err := x509.MarshalECPrivateKey(&wallet.PrivateKey)
		if err != nil {
			panic(err)
		}

		dto[address] = walletDTO{
			PrivateKey: der,
			PublicKey:  wallet.PublicKey,
		}
	}

	encoder := gob.NewEncoder(&content)

	if err := encoder.Encode(dto); err != nil {
		panic(err)
	}

	if err := os.WriteFile(walletFile, content.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	dto := make(map[string]walletDTO)

	decoder := gob.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&dto); err != nil {
		return err
	}

	ws.Wallets = make(map[string]*Wallet)

	for address, w := range dto {
		private, err := x509.ParseECPrivateKey(w.PrivateKey)
		if err != nil {
			return err
		}

		ws.Wallets[address] = &Wallet{
			PrivateKey: *private,
			PublicKey:  w.PublicKey,
		}
	}

	return nil
}