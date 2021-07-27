package database

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"nessaj_proxy/constant"
)

func GetDB() (*badger.DB, error) {
	dir := constant.DBPath

	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, fmt.Errorf("open database failed: %s", err)
	}
	return db, nil
}

/*
@param []byte: key
@param []byte： status
@return error: error msg
*/
func Set(k []byte, v []byte, db *badger.DB) error {
	if err := db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(k, v); err != nil {
			return fmt.Errorf("txn set failed: %s", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("Update transation failed: %s", err)
	}

	return nil
}

/*
@param: []byte: key
@return []byte： status
@return error: error msg
*/
func View(k []byte, db *badger.DB) ([]byte, error) {
	var ret []byte
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return fmt.Errorf("get failed: %s", err)
		}

		if err = item.Value(func(val []byte) error {
			return nil
		}); err != nil {
			return fmt.Errorf("get item value failed: %s", err)
		}

		ret, err = item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("copy value failed: %s", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("veiw transation failed: %s", err)
	}
	return ret, nil
}

/*
@return map[string]string： a map of [ip]status pair
@return int: the total number of records and error msg
@return error: error msg
*/
func ListAll(db *badger.DB) (map[string]string, int, error) {
	res := make(map[string]string)
	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)

		defer it.Close()

		var k string
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k = string(item.Key())
			err := item.Value(func(v []byte) error {
				return nil
			})
			if err != nil {
				return err
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("copy value failed: %s", err)
			}
			res[k] = string(val)
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}
	return res, len(res), nil
}
