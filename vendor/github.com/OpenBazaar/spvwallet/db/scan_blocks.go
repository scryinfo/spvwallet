package db

import (
	"database/sql"
	"fmt"
	"github.com/scryinfo/wallet-interface"
	"sync"
)

type ScanBlocksDB struct {
	db   *sql.DB
	lock *sync.RWMutex
}

func (sbdb *ScanBlocksDB) Get(blockHash string) (wallet.ScanBlock, error) {
	sbdb.lock.RLock()
	defer sbdb.lock.RUnlock()
	var sbs wallet.ScanBlock
	stmt, err := sbdb.db.Prepare("select * from scanBlocks where blockHash=?")
	if err != nil {
		return sbs, err
	}
	defer stmt.Close()
	var isFixScan int
	var blockHeight int
	err = stmt.QueryRow(blockHash).Scan(&blockHeight, &isFixScan)
	if err != nil {
		return sbs, err
	}
	sbs = wallet.ScanBlock{
		BlockHash:   blockHash,
		BlockHeight: blockHeight,
		IsFixScan:   isFixScan,
	}
	return sbs, nil
}

func (sbdb *ScanBlocksDB) Put(blockHash string, blockHeight int, isFixScan int) error {
	sbdb.lock.Lock()
	defer sbdb.lock.Unlock()
	tx, err := sbdb.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert or replace into scanBlocks(blockHash, blockHeight, isFixScan) values(?,?,?)")
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		fmt.Println("err is ", err)
		return err
	}
	_, err = stmt.Exec(blockHash, blockHeight, isFixScan)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sbdb *ScanBlocksDB) UpdateBlock(blockHash string, isFixScan int) error {
	sbdb.lock.Lock()
	defer sbdb.lock.Unlock()
	tx, err := sbdb.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("update scanBlocks set isFixScan=? where blockHash=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(int(isFixScan), blockHash)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sbdb *ScanBlocksDB) Delete(blockHash string) error {
	sbdb.lock.Lock()
	defer sbdb.lock.Unlock()
	_, err := sbdb.db.Exec("delete from scanBlocks where blockHash=?", blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (sbdb *ScanBlocksDB) GetLatestUnScanBlockHash() (string, error) {
	sbdb.lock.RLock()
	defer sbdb.lock.RUnlock()
	var blockHash string
	stmt, err := sbdb.db.Prepare("select scanBlocks.blockHash from scanBlocks where isFixScan=? limit 0,1")
	if err != nil {
		return blockHash, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(0).Scan(&blockHash)
	if err != nil {
		return blockHash, err
	}
	return blockHash, nil
}
