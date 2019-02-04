package main

import (
	. "github.com/protosam/go-libnss/structs"
	flag "github.com/spf13/pflag" // because "flag" is too stupid to follow GNU precedence
	"fmt"
	"os"
	"path"
	"strconv"
	"errors"
)

func main(){
	f_username := flag.String("username", "", "Username (root only)")
	f_password := flag.String("password", "!!", "Password")
	flag.Parse()

	if os.Getuid() == 0 && *f_username == "" {
		fmt.Println("Root is a system user. Use passwd command instead.")
		os.Exit(1)
	}
	
	if os.Getuid() != 0 && *f_username != "" {
		fmt.Println("Only root can use the --username flag.")
		os.Exit(1)
	}

	var shadow_entry Shadow
	found := false
	if *f_username == "" {
		for _, entry := range PasswdDB {
			if entry.UID == uint(os.Getuid()) {
				*f_username = entry.Username
			}
		}
	}else{
		for _, entry := range PasswdDB {
			if entry.Username == *f_username {
				*f_username = entry.Username
			}
		}
	}

	for _, entry := range ShadowDB {
		if entry.Username == *f_username {
			shadow_entry = entry
			found = true
		}
	}

	if !found {
		fmt.Println("Try passwd command instead. Your user is not stored in etcd.")
		os.Exit(1)
	}

	var err error
	hashedword := *f_password
	if *f_password != "!!" {
		hashedword, err = shadow_word(*f_password)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	shadow_entry.Password = hashedword

	resp, err := etcd_update(
		path.Join("/etc/shadow/", *f_username),
		shadow_entry,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("Failed to update entry in etcd:/etc/shadow/"))
		os.Exit(1)
	}
	
	fmt.Println("Updated", *f_username ,"successfully.")
}
