package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
)

func httpGet(address string, path string) (string, error) {
	resp, err := http.Get("http://" + address + "/" + path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//log.Println("DEBUG:Status code:", resp.StatusCode, " path:", path, " address:", address)
	body := []byte{}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	}
	return string(body), nil
}

func getPeers(peerList []string) []string {
	newList := []string{config.myaddress}
	for _, peer := range peerList {
		listString, err := httpGet(peer, "peerlist")
		if err != nil {
			log.Println("Failed to get peerlist from ", peer)
			break
		}
		// Add new peers to the list
		list := strings.Split(listString, "\n")
		for _, newpeer := range list {
			new := true
			for _, peer := range newList {
				if peer == newpeer || peer == ""{
					new = false
					break
				}
			}
			if new {
				newList = append(newList, newpeer)
			}
		}
	}
	return newList
}

func getLatestBlock() Block {
	for _, peer := range peerList.list {
		block, err := httpGet(peer, "getblock")
		if err != nil {
			log.Println("Failed to get peerlist from ", peer)
			break
		}
		//log.Println("DEBUG:latest block body:", block, " address:", peer, " byte slice:", []byte(block))
		b, err := JSONtoBlock(bytes.Trim([]byte(block), "\x00"))
		if err != nil {
			log.Println("Could not get latest block:", err)
		} else {
			return b
		}
	}
	return Block{}
}

func getBlock(hash string) Block {
	for _, peer := range peerList.list {
		block, err := httpGet(peer, "getblock/"+hash)
		if err != nil {
			log.Println("Failed to get block:", hash, " from ", peer)
			break
		}

		b, err := JSONtoBlock(bytes.Trim([]byte(block), "\x00"))
		if err != nil {
			fmt.Println("Could not get block:", hash, " from ", peer, err)
		}
		_, err = b.verifyBlock()
		if err != nil {
			fmt.Println("Could not verify block:", hash, " from ", peer, err)
		} else {
			return b
		}
	}
	return Block{}
}

func advertiseBlock(b Block) {
	data, err := b.blockToJSON()
	if err != nil {
		log.Println("Could not advertise block,", err)
		return
	}
	for _, peer := range peerList.list {
		req, err := http.NewRequest("GET", "http://"+peer+"/recieveblock", bytes.NewReader(data))
		if err != nil {
			log.Println("Could not advertise block to", peer, err)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Could not advertise block to", peer, err)
			return
		}
		log.Println("Sent new block to", peer, resp.Status)
	}
}
