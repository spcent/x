package rest

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

// 基于CSV文件的数据库实现
// 每个记录的第一个字段为ID，第二个字段为版本号
// 版本号从1开始，每次更新版本号加1
// 版本号为0的记录表示删除
type csvDB struct {
	mu      sync.Mutex       // 互斥锁，用于保护读写操作
	f       *os.File         // 文件句柄
	w       *csv.Writer      // CSV写入器
	index   map[string]int64 // ID到文件偏移量的索引
	version map[string]int64 // ID到版本号的索引
}

func NewCSVDB(path string) (*csvDB, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	db := &csvDB{
		f:       f,
		w:       csv.NewWriter(f),
		index:   map[string]int64{},
		version: map[string]int64{},
	}
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	for {
		pos := r.InputOffset()
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(rec) > 0 {
			db.index[rec[0]] = pos
			db.version[rec[0]], _ = strconv.ParseInt(rec[1], 10, 64)
		}
	}

	return db, nil
}

func (db *csvDB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.w.Flush()
	return db.f.Close()
}

func (db *csvDB) append(r Record) error {
	pos, err := db.f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	err = db.w.Write(r)
	if err != nil {
		return err
	}

	db.w.Flush()
	db.index[r[0]] = pos
	db.version[r[0]], err = strconv.ParseInt(r[1], 10, 64)
	return err
}

func (db *csvDB) Create(r Record) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if r[0] == "" || r[1] != "1" || db.version[r[0]] != 0 {
		return errors.New("invalid record")
	}
	return db.append(r)
}

func (db *csvDB) Update(r Record) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if len(r) == 0 || r[1] != strconv.FormatInt(db.version[r[0]]+1, 10) {
		return errors.New("invalid record version")
	}
	return db.append(r)
}

func (db *csvDB) Delete(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.version[id] < 1 {
		return errors.New("record not found")
	}
	return db.append(Record{id, "0"})
}

func (db *csvDB) Get(id string) (Record, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.version[id] < 1 {
		return nil, errors.New("record not found")
	}
	offset, ok := db.index[id]
	if !ok {
		return nil, nil
	}
	if _, err := db.f.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	r := csv.NewReader(db.f)
	rec, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(rec) > 0 && rec[0] != id {
		log.Println(rec)
		return nil, errors.New("corrupted index")
	}
	return rec, nil
}

func (db *csvDB) Iter() func(yield func(Record, error) bool) {
	return func(yield func(Record, error) bool) {
		db.mu.Lock()
		defer db.mu.Unlock()

		if _, err := db.f.Seek(0, io.SeekStart); err != nil {
			yield(nil, err)
			return
		}

		r := csv.NewReader(db.f)
		r.FieldsPerRecord = -1
		for {
			rec, err := r.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				yield(nil, err)
				return
			}
			if len(rec) < 2 {
				continue
			}
			id, version := rec[0], rec[1]
			if version == "0" || version != strconv.FormatInt(db.version[id], 10) {
				continue // deleted items or outdated versions
			}
			if !yield(rec, nil) {
				return
			}
		}
	}
}
