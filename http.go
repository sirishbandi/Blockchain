package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func httpGet(adress string, path string) (string, error) {
	resp, err := http.Get("http://" + adress + "/" + path)
	if err != nil {
		return "", err
	}
	body := []byte{}
	if _, err = resp.Body.Read(body); err != nil {
		return "", err
	}
	return string(body), nil
}

func getPeers(peerList []string) []string {
	for i, peer := range peerList {
		for retry := RETRY_COUNT; retry > 0; retry-- {
			listString, err := httpGet(peer, "peerlist")
			if err != nil {
				log.Println("Failed to get peerlist from ", peer, ".Retry=", retry)
				time.Sleep(5 * time.Second)
				continue
			}

			// Counld not get a list of peers, exit.
			if retry == 1 {
				log.Println("Could not get peerlist from ", peer, ". Deleting peer.")
				peerList = append(peerList[:i], peerList[i+1:]...)
			}

			// Add new peers to the list
			list := strings.Split(listString, "\n")
			for _, newpeer := range list {
				new := true
				for _, peer := range peerList {
					if peer == newpeer {
						new = false
						break
					}
				}
				if new {
					peerList = append(peerList, newpeer)
				}
			}
		}
	}
	return peerList
}

func getLatestBlock() Block {
	for _, peer := range peerList.list {
		block, err := httpGet(peer, "getblock/")
		if err != nil {
			log.Println("Failed to get peerlist from ", peer)
			break
		}

		b, err := JSONtoBlock([]byte(block))
		if err != nil {
			fmt.Println("Could not get latest block:", err)
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

		b, err := JSONtoBlock([]byte(block))
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
