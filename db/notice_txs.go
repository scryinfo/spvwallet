package db

import (
	"database/sql"
	"fmt"
	"github.com/scryinfo/wallet-interface"
	"sync"
)

type NoticeTxsDB struct {
	db   *sql.DB
	lock *sync.RWMutex
}

func (txdb *NoticeTxsDB) Get(txHash string) (wallet.NoticeTx, error) {
	txdb.lock.RLock()
	defer txdb.lock.RUnlock()
	var noTx wallet.NoticeTx
	stmt, err := txdb.db.Prepare("select * from noticeTx where txHash=?")
	if err != nil {
		return noTx, err
	}
	defer stmt.Close()
	var value int
	var wechatTxId string
	var isNotice int
	err = stmt.QueryRow(txHash).Scan(&value, &wechatTxId, &isNotice)
	if err != nil {
		return noTx, err
	}
	noTx = wallet.NoticeTx{
		TxHash:     txHash,
		Value:      value,
		WechatTxId: wechatTxId,
		IsNotice:   isNotice,
	}
	return noTx, nil
}

func (txdb *NoticeTxsDB) Put(txHash string, value int, wechatTxId string, isNotice int) error {
	txdb.lock.Lock()
	defer txdb.lock.Unlock()
	tx, err := txdb.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert or replace into noticeTx(txHash, value, wechatTxId,isNotice) values(?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		fmt.Println("err is ", err)
		return err
	}
	_, err = stmt.Exec(txHash, value, wechatTxId, isNotice)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (txdb *NoticeTxsDB) UpdateBlock(txHash string, value int, wechatTxId string, isNotice int, noticedCount int) error {
	txdb.lock.Lock()
	defer txdb.lock.Unlock()
	tx, err := txdb.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("update noticeTx set value=?, wechatTxId=? , isNotice=? ,noticedCount=? where txHash=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(value, wechatTxId, isNotice, noticedCount, txHash)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (txdb *NoticeTxsDB) Delete(txHash string) error {
	txdb.lock.Lock()
	defer txdb.lock.Unlock()
	_, err := txdb.db.Exec("delete from noticeTx where txHash=?", txHash)
	if err != nil {
		return err
	}
	return nil
}
