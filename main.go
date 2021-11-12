package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func readFlags() {
	flag.StringVar(&config.user, "user", "aaaaaaaaaa", "My public user key.")
	flag.StringVar(&config.address, "address", "127.0.0.1:8080", "The address of a peer.")
	flag.StringVar(&config.myaddress, "myaddress", "127.0.0.1:8080", "Myaddress")
	flag.BoolVar(&config.init, "init", false, "Create a new chain if true")
	flag.Parse()
}

func readFile(file string) ([]byte, error) {
	f, err := os.Open("blocks/" + file + ".json")
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	data := make([]byte, MAX_BLOCK_SIZE)
	_, err = f.Read(data)
	return data, err
}

func writeFile(file string, data []byte) error {
	f, err := os.Create("blocks/" + file + ".json")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func createInitBlock() {
	err := os.MkdirAll("blocks", 0755)
	if err != nil {
		log.Fatal(err)
	}
	b := Block{
		BlockID:       0,
		PrevBlockHash: "",
		Nounce:        0,
		User:          config.user,
		Data:          "",
	}

	data, err := b.blockToJSON()
	if err != nil {
		log.Fatal(err)
	}
	hash, err := b.calcualteHash()
	if err != nil {
		log.Fatal(err)
	}

	blockState = BlockState{
		currBlockHash: hash,
		block: Block{
			BlockID:       b.BlockID + 1,
			PrevBlockHash: hash,
		},
	}

	// Starting Block
	err = writeFile("1stblock", data)
	if err != nil {
		log.Fatal(err)
	}

	// Also create a block with hash for reading.
	err = writeFile(hash, data)
	if err != nil {
		//log.Fatal(err)
		log.Println("Hash did not match, but continuing hoping it was 1st block")
	}
}

func pullAllBlocks() {
	err := os.MkdirAll("blocks", 0755)
	if err != nil {
		log.Fatal(err)
	}

	b := getLatestBlock()
	err = blockState.syncBlocks(b)
	if err != nil {
		log.Fatal(err)
	}
}

func bootstrap() {
	readFlags()
	peerList.list = []string{config.address}
	// Get the list of Peers from my negibour
	// peerList.updatePeers() if run it would remove self from list as server is not yet running
	log.Println("Got the following peers from seed:", peerList.list)

	//Create the 1st block or read blocks from peer to be upto date
	if config.init {
		log.Println("Create a new first block...")
		createInitBlock()
	} else {
		log.Println("Syncing blocks from peers..")
		pullAllBlocks()
	}

}

func main() {
	log.Println("Starting Blockchain node:")
	bootstrap()
	log.Println("Bootstap Complete. setting up hasher and HTTP server")
	dataQ = make(chan string, MAX_BACKLOG)
	syncQ = make(chan Block, MAX_BACKLOG)
	go func() {
		time.Sleep(5 * time.Second) //Wait 5sec to ensure HTTP server starts.
		peerList.cronUpdatePeer()
	}()
	go blockState.blockOperations()
	httpsertup()
	log.Println("Starting HHTP Server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln("Could not start HTTP server,", err)
	}
}
