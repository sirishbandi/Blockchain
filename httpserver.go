package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func httpsertup() {
	http.HandleFunc("/peerlist", peerListFunc)
	http.HandleFunc("/getblock", getBlockFunc)
	http.HandleFunc("/getblock/", getBlockFunc)
	http.HandleFunc("/addblock", addBlockFunc)
	http.HandleFunc("/recieveblock", recieveBlockFunc)
}

func peerListFunc(w http.ResponseWriter, req *http.Request) {
	t := strings.Split(req.RemoteAddr, ":")
	ip := t[0] + ":8080"
	
	new := true
	for _, peer := range peerList.list {
		if peer == ip{
			new = false
		}
		// We use newline as delimiter
		fmt.Fprintln(w, peer)
	}
	if new {
		log.Println("Got new Peer for /peerlist,", ip)
		peerList.lock.Lock()
		peerList.list = append(peerList.list, ip)
		peerList.lock.Unlock()
	}
}

func getBlockFunc(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	path := strings.Split(url, "/")
	hash := blockState.currBlockHash
	// If a hash is passed in request URL path we use it else default to lastest block
	if len(path) > 2 {
		hash = path[2]
	}
	data, err := readFile(hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not read file,", err)
		return
	}
	fmt.Fprint(w, string(data))
}

func addBlockFunc(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not read body,", err)
		return
	}
	select {
	case dataQ <- string(body):
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Q full,", err)
		return
	}
}

func recieveBlockFunc(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not read body when receving block,", err)
		return
	}
	b, err := JSONtoBlock(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not convert body to block when receving block,", err)
		return
	}

	_, err = b.verifyBlock()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, blockState.currBlockHash)
		return
	}

	// Is the blockstates match, then dont do anything
	if blockState.currBlockID == b.BlockID {
		return
	}

	// The block held by this node is greater than the block advertised. Send back the newest block.
	if blockState.currBlockID > b.BlockID {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, blockState.currBlockHash)
		return
	}

	// If the revieved new block is newer than the current block we will need to sync all missing blocks.
	// Need to run only one sync block at a time.
	syncQ <- b
}
