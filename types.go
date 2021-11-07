package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

const (
	RETRY_COUNT    = 3
	MAX_BLOCK_SIZE = 1 * 1024 // in bytes
	HASH_THESHOLD  = "000000f03fd5c584d35b610a6a75467460509153360492992deea037978fd329192ce66c51a9e635771f5874de0562a14665eeb8a7294edb9224481878e0b5db"
	HASH_MAX_TIME  = 5 * time.Minute
	MAX_BACKLOG    = 5
	PEER_SYNC_TIME = 1 * time.Minute
	FIRST_BLOCK = "40198a17205aca94791e2425fda69eb0b6af99626fe17159c82426b2a14ca3653baa999a66d07dd977b58358607b99d546249f08b1448d6c4fd70389f6bf845e"
)

type Config struct {
	user    string
	address string
	init    bool
}

var config Config

type PeerList struct {
	list []string
	lock sync.Mutex
}

var peerList PeerList

type Block struct {
	BlockID       int64  `json:"blockId"`
	PrevBlockHash string `json:"prevBlockHash"`
	Nounce        int64  `json:"nounce"`
	User          string `json:"user"`
	Data          string `json:"data"`
}

type BlockState struct {
	currBlockHash string
	currBlockID   int64
	block         Block
	lock          sync.Mutex
}

var blockState BlockState
var dataQ chan string
var syncQ chan Block

// Functions

func (b Block) blockToJSON() ([]byte, error) {
	return json.Marshal(b)
}

func JSONtoBlock(b []byte) (Block, error) {
	var block Block
	err := json.Unmarshal(b, &block)
	return block, err
}

func (b Block) calcualteHash() (string, error) {
	data, err := b.blockToJSON()
	if err != nil {
		return "", err
	}
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:]), nil
}
