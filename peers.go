package main

import (
	"time"
	"log"
)

func (peerList *PeerList) updatePeers() {
	tempList := make([]string,1)
	// Deep copy below, not using copy() to remove newline
	tempList[0] = peerList.list[0]
	for i:=1; i<len(peerList.list); i++ {
		if peerList.list[i] == "" {break }
		tempList = append(tempList,peerList.list[i])
	}
	log.Println("Updating peerList, current list:", tempList)
	list := getPeers(tempList)
	peerList.lock.Lock()
	peerList.list = list
	peerList.lock.Unlock()
	log.Println("Updated peerList, new list:", peerList)
}

func (PeerList *PeerList) cronUpdatePeer() {
	for {
		time.Sleep(PEER_SYNC_TIME)
		peerList.updatePeers()
	}
}
