package main

import (
	"time"
	"log"
)

func (peerList *PeerList) updatePeers() {
	tempList := []string{}
	copy(tempList, peerList.list)
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
