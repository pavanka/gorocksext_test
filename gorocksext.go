package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pavanka/gorocksext"
	"github.com/tecbot/gorocksdb"
)

func read_all(db_dir string) error {
	dbOpts := gorocksdb.NewDefaultOptions()
	db, err := gorocksdb.OpenDb(dbOpts, db_dir)
	if err != nil {
		return err
	}
	defaultRO := gorocksdb.NewDefaultReadOptions()
	iter := db.NewIterator(defaultRO)
	iter.SeekToFirst()
	for iter.Valid() {
		key := iter.Key()
		defer key.Free()
		fmt.Println(string(key.Data()))
		iter.Next()
	}
	db.Close()
	return nil
}

func check_checkpoints() error {
	dbOpts := gorocksdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)
	if err := os.RemoveAll("/tmp/checkpoint_db"); err != nil {
		return err
	}

	db, err := gorocksdb.OpenDb(dbOpts, "/tmp/checkpoint_db")
	if err != nil {
		return err
	}
	writeOpts := gorocksdb.NewDefaultWriteOptions()
	for i := 0; i < 16; i++ {
		key := []byte(fmt.Sprint(i))
		if err := db.Put(writeOpts, key, key); err != nil {
			return err
		}
		if i == 8 {
			if err := os.RemoveAll("/tmp/checkpoint_db1"); err != nil {
				return err
			}
			gorocksext.CreateCheckpoint(db, "/tmp/checkpoint_db1")
		}
	}
	db.Close()

	fmt.Println("Full Db")
	if err := read_all("/tmp/checkpoint_db"); err != nil {
		return err
	}
	fmt.Println("Checkpointed snapshot")
	if err := read_all("/tmp/checkpoint_db1"); err != nil {
		return err
	}
	return nil
}

func write_multi_cfs() error {
	dbOpts := gorocksdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)
	if err := os.RemoveAll("/tmp/multicf_db"); err != nil {
		return err
	}

	db, err := gorocksdb.OpenDb(dbOpts, "/tmp/multicf_db")
	if err != nil {
		return err
	}

	var handles []*gorocksdb.ColumnFamilyHandle
	for i := 0; i < 4; i++ {
		handle, err := db.CreateColumnFamily(dbOpts, fmt.Sprint(i))
		if err != nil {
			return err
		}
		handles = append(handles, handle)
	}

	writeOpts := gorocksdb.NewDefaultWriteOptions()
	if err := db.Put(writeOpts, []byte("default"), []byte("default")); err != nil {
		return err
	}
	for i := 0; i < 16; i++ {
		key := []byte(fmt.Sprint(i))
		if err := db.PutCF(writeOpts, handles[i%4], key, key); err != nil {
			return err
		}
	}
	db.Close()
	return nil
}

func read_multi_cfs() error {
	dbOpts := gorocksdb.NewDefaultOptions()
	defaultRO := gorocksdb.NewDefaultReadOptions()
	db, handles, err := gorocksdb.OpenDbColumnFamilies(
		dbOpts,
		"/tmp/multicf_db",
		[]string{"default", "0", "1", "2", "3"},
		[]*gorocksdb.Options{dbOpts, dbOpts, dbOpts, dbOpts, dbOpts},
	)
	iters, err := gorocksext.NewIterators(defaultRO, db, handles)
	if err != nil {
		return err
	}
	for i, iter := range iters {
		fmt.Printf("COUTING FOR ITER: %d\n", i)
		iter.SeekToFirst()
		for iter.Valid() {
			fmt.Println(string(iter.Key().Data()))
			defer iter.Key().Free()
			iter.Next()
		}
	}
	db.Close()
	return nil
}

func create_purge_backups() error {
	dbOpts := gorocksdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)
	if err := os.RemoveAll("/tmp/backups_db"); err != nil {
		return err
	}

	db, err := gorocksdb.OpenDb(dbOpts, "/tmp/backups_db")
	if err != nil {
		return err
	}

	if err := os.RemoveAll("/tmp/db_backups"); err != nil {
		return err
	}
	be, err := gorocksdb.OpenBackupEngine(dbOpts, "/tmp/db_backups")
	if err != nil {
		return err
	}

	writeOpts := gorocksdb.NewDefaultWriteOptions()
	for i := 0; i < 5; i++ {
		key := []byte(fmt.Sprint(i))
		if err := db.Put(writeOpts, key, key); err != nil {
			return err
		}
		if err := be.CreateNewBackup(db); err != nil {
			return err
		}
	}
	fmt.Println("Num available initially: ", be.GetInfo().GetCount())
	if err := gorocksext.PurgeOldBackups(be, 4); err != nil {
		return err
	}
	fmt.Println(be.GetInfo().GetCount())
	if err := gorocksext.PurgeOldBackups(be, 2); err != nil {
		return err
	}
	fmt.Println(be.GetInfo().GetCount())
	if err := gorocksext.PurgeOldBackups(be, 0); err != nil {
		return err
	}
	fmt.Println(be.GetInfo().GetCount())

	db.Close()
	return nil
}

func main() {
	fmt.Println("\ncheck multiple checkpoints")
	if err := check_checkpoints(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nwrite to multiple cfs")
	if err := write_multi_cfs(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nreading from multiple cfs")
	if err := read_multi_cfs(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\ncreate and purge rocksdb backups")
	if err := create_purge_backups(); err != nil {
		log.Fatal(err)
	}
}
