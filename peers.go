package main

import "time"

func (peerList *PeerList) updatePeers() {
	tempList := []string{}
	copy(tempList, peerList.list)
	list := getPeers(tempList)
	peerList.lock.Lock()
	peerList.list = list
	peerList.lock.Unlock()
}

func (PeerList *PeerList) cronUpdatePeer() {
	for {
		time.Sleep(PEER_SYNC_TIME)
		peerList.updatePeers()
	}
}
