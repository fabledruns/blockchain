package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/fabledruns/blockchain/blockchain"
	"github.com/fabledruns/blockchain/wallet"
)

type CommandLine struct {}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage: ")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" createblockchain -address ADDRESS - create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println(" printchain - print all the blocks of the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - send amount of coins")
	fmt.Println(" createwallet - creates a new wallet")
	fmt.Println(" listaddresses - lists the addresses in our wallet file")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses  := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address: %s\n", address)
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)

		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string) {
	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished")
}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		fmt.Println("Address is not valid")
	}
	chain := blockchain.ContinueBlockChain(address)
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash) - 4]
	UTXOs := chain.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		fmt.Println("Address is not valid")
		return
	}
	if !wallet.ValidateAddress(to) {
		fmt.Println("Address is not valid")
		return
	}
	chain := blockchain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})

	fmt.Printf("Success!")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockChainAddress := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
		case "getbalance":
			err := getBalanceCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "createblockchain":
			err := createBlockChainCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "send":
			err := sendCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "print":
			err := printChainCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "createwallet":
			err := createWalletCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "listaddresses":
			err := listAddressesCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		default:
			cli.printUsage()
			runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			createBlockChainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockChainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
}