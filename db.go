package main

import (
  "time"
  "github.com/boltdb/bolt"
)

// CreateDB create a database file if it if was not exist
func CreateDB() error {
  db, err := bolt.Open("MusicBot.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists([]byte("ChannelDB"))
    if err != nil {
      return err
    }
    return nil
  })
  return err
}

// PutDB ignore o unignore a test channel
func PutDB(channelID, ignored string) error {
  db, err := bolt.Open("MusicBot.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    return err
  }
  defer db.Close()

  db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("ChannelDB"))
    err := b.Put([]byte(channelID), []byte(ignored))
    return err
  })
  return err
}

// GetDB read if a text channel is ignored
func GetDB(channelID string) string {
  var v []byte
  db, err := bolt.Open("MusicBot.db", 0600, &bolt.Options{ReadOnly: true})
  if err != nil {
    return ""
  }
  defer db.Close()

  db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("ChannelDB"))
    v = b.Get([]byte(channelID))
    return nil
  })
  return string(v)
}
