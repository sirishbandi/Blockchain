package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

func (b *Block) verifyBlock() (bool, error) {
	hash, err := b.calcualteHash()
	if err != nil {
		return false, err
	}

	if hash > HASH_THESHOLD {
		return false, errors.New(fmt.Sprintln("New block ", b, "is not having correct hash, currblockhash", blockState.currBlockHash))
	}
	return true, nil
}

func (blockState *BlockState) verifyBlock(b Block) (bool, error) {
	var err error
	_, err = b.verifyBlock()
	if err != nil {
		return false, err
	}

	if blockState.currBlockHash != b.PrevBlockHash {
		return false, errors.New(fmt.Sprintln("New block ", b, "is not having correct prevblockhash field, currblockhash", blockState.currBlockHash))
	}

	if blockState.currBlockID+1 != b.BlockID {
		return false, errors.New(fmt.Sprintln("New block ", b, "is not having correct blockid field, currblockid", blockState.currBlockID))
	}
	return true, nil
}

func findMatchingHash(ctx context.Context, b *Block) {
	var count int64
	begin := time.Now()
	h, err := b.calcualteHash()
	if err != nil {
		log.Println("Could not get hash,", err)
		return
	}
	for h > HASH_THESHOLD {
		count++
		b.Nounce = b.Nounce + 1
		h, err = b.calcualteHash()
		if err != nil {
			log.Println("Could not get hash,", err)
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
	log.Println("Found a good nounce after ", count, " hashes, took time:", time.Since(begin))
}

func (blockState *BlockState) updateBlock(b Block) error {
	var err error
	blockState.lock.Lock()
	defer blockState.lock.Unlock()
	blockState.currBlockHash, err = b.calcualteHash()
	if err != nil {
		return err
	}
	blockState.currBlockID = b.BlockID
	blockState.block = b
	return nil
}

func (blockState *BlockState) blockOperations() {
	var cancelFunc context.CancelFunc
	var ctx context.Context
	wg := sync.WaitGroup{}
	for {
		select {
		case data := <-dataQ:
			//*********************************
			// TODO: This wait force the running addBlock to complete, in this time all sync from peers are ignored. Need to fix this.
			wg.Wait()
			b := Block{
				BlockID:       blockState.currBlockID + 1,
				PrevBlockHash: blockState.currBlockHash,
				Nounce:        0,
				User:          config.user,
				Data:          data,
			}
			ctx, cancelFunc = context.WithCancel(context.Background())
			// To force only one addblock function to run at a time
			wg.Add(1)
			go func() {
				blockState.addBlock(ctx, b)
				defer wg.Done()
			}()

		case b := <-syncQ:
			if blockState.currBlockID >= b.BlockID {
				log.Println("synch failed: Block trying to be synced is too old")
				continue
			}
			// Stop the ongoing hash
			if cancelFunc != nil {
				cancelFunc()
			}
			// Get up to date on the blocks
			err := blockState.syncBlocks(b)
			if err != nil {
				log.Println("Error syncing chain:", err)
			}
		}
	}
}

func (blockState *BlockState) addBlock(ctx context.Context, newBlock Block) {
	tctx, cancelFunc := context.WithTimeout(ctx, HASH_MAX_TIME)
	findMatchingHash(tctx, &newBlock)
	cancelFunc()
	if v, err := blockState.verifyBlock(newBlock); !v {
		log.Println("Could not add data ", newBlock, " to the chain.", err)
		return
	}

	err := blockState.updateBlock(newBlock)
	if err != nil {
		log.Println("Could not add data ", newBlock, " to the chain.", err)
		return
	}
	fileData, err := blockState.block.blockToJSON()
	if err != nil {
		log.Println("Could not add data ", newBlock, " to the chain.", err)
		return
	}
	err = writeFile(blockState.currBlockHash, fileData)
	if err != nil {
		log.Println("Could not add data ", newBlock, " to the chain.", err)
		return
	}
	go advertiseBlock(newBlock)
}

func (blockState *BlockState) syncBlocks(b Block) error {
	_, err := b.verifyBlock()
	if err != nil {
		return errors.New("synch failed:" + err.Error())
	}

	hash, err := b.calcualteHash()
	if err != nil {
		return errors.New("synch failed:" + err.Error())
	}

	// Latest Block
	data, err := b.blockToJSON()
	if err != nil {
		return errors.New("synch failed:" + err.Error())
	}
	err = writeFile(hash, data)
	if err != nil {
		return errors.New("synch failed:" + err.Error())
	}

	// Get all other blocks
	for blockHash := b.PrevBlockHash; blockHash != FIRST_BLOCK; {
		b := getBlock(hash)
		data, err := b.blockToJSON()
		if err != nil {
			return err
		}
		err = writeFile(hash, data)
		if err != nil {
			return err
		}
		blockHash = b.PrevBlockHash
	}
	// b is still the newest block
	blockState.updateBlock(b)
	return nil
}
