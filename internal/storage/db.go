package storage

import (
    "time"

    "go.etcd.io/bbolt"
)

type DB struct {
    db *bbolt.DB
}

func New(path string) (*DB, error) {
    db, err := bbolt.Open(path, 0600, nil)
    if err != nil {
        return nil, err
    }
    err = db.Update(func(tx *bbolt.Tx) error {
        tx.CreateBucketIfNotExists([]byte("agents"))
        tx.CreateBucketIfNotExists([]byte("commands"))
        tx.CreateBucketIfNotExists([]byte("results"))
        return nil
    })
    if err != nil {
        return nil, err
    }
    return &DB{db: db}, nil
}

func (d *DB) Close() error {
    return d.db.Close()
}

func (d *DB) SaveAgent(peerID string, t time.Time) {
    d.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("agents"))
        return b.Put([]byte(peerID), []byte(t.Format(time.RFC3339)))
    })
}

func (d *DB) QueueCommand(agentID, cmd string) {
    d.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("commands"))
        key := []byte(agentID + "-" + time.Now().Format(time.RFC3339Nano))
        return b.Put(key, []byte(cmd))
    })
}

func (d *DB) PopCommand(agentID string) (string, error) {
    var cmd string
    err := d.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("commands"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            // check prefix
            if len(k) >= len(agentID) && string(k[:len(agentID)]) == agentID {
                cmd = string(v)
                return b.Delete(k)
            }
        }
        return nil
    })
    return cmd, err
}

func (d *DB) SaveResult(agentID, res string) {
    d.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("results"))
        key := []byte(agentID + "-" + time.Now().Format(time.RFC3339Nano))
        return b.Put(key, []byte(res))
    })
}

func (d *DB) PopResult(agentID string) (string, error) {
    var res string
    err := d.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("results"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            if len(k) >= len(agentID) && string(k[:len(agentID)]) == agentID {
                res = string(v)
                return b.Delete(k)
            }
        }
        return nil
    })
    return res, err
}
