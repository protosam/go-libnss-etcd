package main

import (
	"fmt"
	. "github.com/protosam/go-libnss/structs"
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/clientv3util"
	"time"
	"encoding/json"

	"os"
	"io/ioutil"
	
	"sort"
	"github.com/kless/osutil/user/crypt/sha512_crypt"
	"crypto/rand"
	"encoding/hex"
)

var etcd_client *clientv3.Client
var etcd_client_success bool

var PasswdDB []Passwd
var GroupDB []Group
var ShadowDB []Shadow

var config ConfigFile

type ConfigFile struct {
		Endpoints []string
		DialTimeout int
		Username string
		Password string
		MinXID int
}

func init(){
	var err error

	// Get relevant config
	var config_file string
	
	
	if os.Geteuid() == 0 {
		config_file = "/etc/nss-etcd-root.conf"
	}else{
		config_file = "/etc/nss-etcd.conf"
	}

	// load in the config
	rawconfig, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Println("Failed to read file:", config_file)
		os.Exit(1)
	}
	
	// make the configuration json useable...
	err = json.Unmarshal(rawconfig, &config)
	if err != nil {
		fmt.Println("Failed to parse file:", config_file)
		os.Exit(1)
	}


	// Setup our etcd client
	etcd_client, err = clientv3.New(clientv3.Config{
		Endpoints:	config.Endpoints,
		DialTimeout:	time.Duration(config.DialTimeout) * time.Second,
		Username:	config.Username,
		Password:	config.Password,
	})
	
	if err != nil {
		etcd_client_success = false
	}else{
		etcd_client_success = true
		PasswdDB = etcd_PasswdDB()
		GroupDB = etcd_GroupDB()
		if os.Geteuid() == 0 {
			ShadowDB = etcd_ShadowDB()
		}
	}
}


func etcd_PasswdDB() []Passwd {
	resp, err := etcd_client.Get(context.Background(), "/etc/passwd/", clientv3.WithPrefix())
	if err != nil {
		return []Passwd{}
	}
	var entries []Passwd
	for _, raw_entry := range resp.Kvs {
		var entry Passwd
		json.Unmarshal(raw_entry.Value, &entry)
		entries = append(entries, entry)
	}
	return entries
}

func etcd_GroupDB() []Group {
	resp, err := etcd_client.Get(context.Background(), "/etc/group/", clientv3.WithPrefix())
	if err != nil {
		return []Group{}
	}

	var entries []Group
	for _, raw_entry := range resp.Kvs {
		var entry Group
		json.Unmarshal(raw_entry.Value, &entry)
		entries = append(entries, entry)
	}
	return entries
}

func etcd_ShadowDB() []Shadow {
	resp, err := etcd_client.Get(context.Background(), "/etc/shadow/", clientv3.WithPrefix())
	if err != nil {
		return []Shadow{}
	}

	var entries []Shadow
	for _, raw_entry := range resp.Kvs {
		var entry Shadow
		json.Unmarshal(raw_entry.Value, &entry)
		entries = append(entries, entry)
	}
	return entries
}

func etcd_insert(key string, value_i interface{}) (*clientv3.TxnResponse, error) {
	value, err := json.Marshal(value_i)
	if err != nil {
		return &clientv3.TxnResponse{}, err
	}

	return etcd_client.Txn(context.Background()).
							If(clientv3util.KeyMissing(key)).
							Then(clientv3.OpPut(key, string(value))).
							Commit()
}

func etcd_delete(key string) (*clientv3.DeleteResponse, error) {
	return etcd_client.Delete(context.Background(), key)
}

func etcd_update(key string, value_i interface{}) (*clientv3.TxnResponse, error) {
	value, err := json.Marshal(value_i)
	if err != nil {
		return &clientv3.TxnResponse{}, err
	}

	return etcd_client.Txn(context.Background()).
							If(clientv3util.KeyExists(key)).
							Then(clientv3.OpPut(key, string(value))).
							Commit()
}






func contains(s []string, b string) bool {
	for _, a := range s {
		if a == b {
			return true
		}
	}
	return false
}

func contains_int(s []int, b int) bool {
	for _, a := range s {
		if a == b {
			return true
		}
	}
	return false
}

func nextid(ids []int) int {
	if len(ids) == 0 {
		return config.MinXID
	}
	sort.Ints(ids)
	lastid := ids[len(ids)-1] + 1
	for id := config.MinXID; id < lastid; id++ {
		if !contains_int(ids, id) {
			return id
		}
	}
	return lastid
}

func getnextid() int {
	var ids []int
	for _, entry := range PasswdDB {
		if !contains_int(ids, int(entry.UID)) {
			ids = append(ids, int(entry.UID))
		}
	}
	for _, entry := range GroupDB {
		if !contains_int(ids, int(entry.GID)) {
			ids = append(ids, int(entry.GID))
		}
	}
	return nextid(ids)
}

func user_exists(username string) bool {
	for _, entry := range PasswdDB {
		if entry.Username == username {
			return true
		}
	}
	return false
}

func get_user(username string) Passwd {
	for _, entry := range PasswdDB {
		if entry.Username == username {
			return entry
		}
	}
	return Passwd{}
}

func group_exists(groupname string) bool {
	for _, entry := range GroupDB {
		if entry.Groupname == groupname {
			return true
		}
	}
	return false
}

func get_group(groupname string) Group {
	for _, entry := range GroupDB {
		if entry.Groupname == groupname {
			return entry
		}
	}
	return Group{}
}

func shadow_word(password string) (string, error) {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	shadow_word := hex.EncodeToString(b)

	c := sha512_crypt.New()
	hash, err := c.Generate([]byte(password), []byte("$6$" + shadow_word))

	return hash, err
}
