package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/OpenBazaar/spvwallet"
	"github.com/OpenBazaar/spvwallet/db"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/scryinfo/scryg/sutils/ssignal"
)

const (
	AddressesPath = "addresses.json"
)

func main() {
	// Create a new config
	config := spvwallet.NewDefaultConfig()

	config.Params = &chaincfg.MainNetParams
	config.Mnemonic = "now inflict diamond try shrimp whip deposit collect such symbol latin cinnamon"
	config.CreationDate = time.Now()

	/*// Make the logging a little prettier
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	formatter := logging.MustStringFormatter(`%{color:reset}%{color}%{time:15:04:05.000} [%{shortfunc}] [%{level}] %{message}`)
	stdoutFormatter := logging.NewBackendFormatter(backend, formatter)
	config.Logger = logging.MultiLogger(stdoutFormatter)

	// Use testnet
	config.Params = &chaincfg.TestNet3Params*/

	_, ferr := os.Stat(config.RepoPath)
	fmt.Println("config.RepoPath is ", config.RepoPath)
	fmt.Println("ferr is ", ferr)
	if os.IsNotExist(ferr) {
		err := os.Mkdir(config.RepoPath, os.ModePerm)
		fmt.Println("Mkdir err is ", err)
	}

	{
		var fileLogFormat = logging.MustStringFormatter(`%{time:15:04:05.000} [%{shortfunc}] [%{level}] %{message}`)

		w := &lumberjack.Logger{
			Filename:   path.Join(config.RepoPath, "logs", "bitcoin.log"),
			MaxSize:    10, // Megabytes
			MaxBackups: 300,
			MaxAge:     30, // Days
		}
		bitcoinFile := logging.NewLogBackend(w, "", 0)
		bitcoinFileFormatter := logging.NewBackendFormatter(bitcoinFile, fileLogFormat)
		config.Logger = logging.MultiLogger(logging.MultiLogger(bitcoinFileFormatter))
	}

	// Select wallet datastore
	sqliteDatastore, _ := db.Create(config.RepoPath)
	config.DB = sqliteDatastore

	// Create the wallet
	wallet, err := spvwallet.NewSPVWallet(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start it!
	wallet.Start()

	{
		var addressList = loadWatchAddressFromFile()
		for _, addr := range addressList {
			addr, err := btcutil.DecodeAddress(addr, config.Params)
			err = wallet.AddWatchedAddresses(addr)
			if err != nil {
				fmt.Println("AddWatchedAddresses(), error is ===>", err)
			} else {
				fmt.Println("AddWatchedAddresses is ===>", addr)
			}
		}
	}

	ssignal.WaitCtrlC(func(s os.Signal) bool { //third wait for exit
		return false
	})
}

func loadWatchAddressFromFile() []string {
	var watchAddressList []string
	ex, err := os.Executable()
	if err == nil {
		ex = filepath.Dir(ex)
	} else {
		ex = ""
	}
	configFile := filepath.Join(ex, AddressesPath)
	fmt.Println("configFile=====>", configFile)
	var resultString = ""
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		fmt.Println("file err=====>", err)
	}
	buf := bufio.NewReader(file)
	for {
		s, err := buf.ReadString('\n')
		resultString += s
		if err != nil {
			if err == io.EOF {
				fmt.Println("Read is ok")
				break
			} else {
				fmt.Println("Error:", err)
				return watchAddressList
			}
		}
	}
	type AddrModel struct {
		AddrList []string `json:"addrlist"`
	}
	var addrModel AddrModel
	err = json.Unmarshal([]byte(resultString), &addrModel)
	if err != nil {
		fmt.Println("Error is", err)
		return watchAddressList
	}
	fmt.Println("addrModel=>", addrModel)
	for _, value := range addrModel.AddrList {
		watchAddressList = append(watchAddressList, value)
	}
	return watchAddressList
}
