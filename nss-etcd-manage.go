package main

import (
	. "github.com/protosam/go-libnss/structs"
	flag "github.com/spf13/pflag" // because "flag" is too stupid to follow GNU precedence
	"os"
	"errors"
	"fmt"
	"path"
	"strconv"
	"go.etcd.io/etcd/clientv3"
)

func main(){
	if os.Getuid() != 0 {
		fmt.Println("You must be root to use this command.")
		os.Exit(1)
	}
	if !etcd_client_success {
		fmt.Println("Failed to connect to etcd service.")
		os.Exit(1)
	}
	argv := os.Args
	if len(argv) == 2 && argv[1] == "help" {
		fmt.Println("See:")
		fmt.Println(argv[0], "user add --help")
		fmt.Println(argv[0], "user delete --help")
		fmt.Println(argv[0], "group add --help")
		fmt.Println(argv[0], "group delete --help")
		fmt.Println(argv[0], "group add-member --help")
		fmt.Println(argv[0], "group remove-member --help")
	}

	if len(argv) < 3 {
		fmt.Println(errors.New("Not enough arguments provided. Run '" + argv[0] + " help' for command line options."))
		os.Exit(1)
	}

	action := argv[1] + " " + argv[2]
	switch action {
	case "user add": user_add()
	case "user delete": user_delete()
	case "group add": group_add()
	case "group delete": group_delete()
	case "group add-member": group_add_member()
	case "group remove-member": group_remove_member()
	default: panic(errors.New("Not enough arguments provided. Run '" + argv[0] + " help' for command line options."))
	}
}

func user_add(){
	// user add --username="testuser" --password="password" --uid=1500 --gid=1500 --comment="Is stored in etcd." --homedir="/home/testuser" --shell="/bin/bash"
	f_username := flag.String("username", "", "Username")
	f_password := flag.String("password", "!!", "Password")
	f_uid := flag.Int("uid", getnextid(), "UID")
	f_gid := flag.Int("gid", getnextid(), "GID")
	f_comment := flag.String("comment", "", "Comment")
	f_homedir := flag.String("homedir", "", "Home directory")
	f_shell := flag.String("shell", "/bin/bash", "Shell")
	flag.Parse()

	if user_exists(*f_username) {
		fmt.Println(errors.New("User " + *f_username + " already exists."))
		os.Exit(1)
	}

	if *f_homedir == "" {
		*f_homedir = path.Join("/home/", *f_username)
	}

	var resp *clientv3.TxnResponse
	var err error

	hashedword := *f_password
	if *f_password != "!!" {
		hashedword, err = shadow_word(*f_password)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	resp, err = etcd_insert(
		path.Join("/etc/passwd/", strconv.Itoa(int(*f_uid))),
		Passwd{
			Username:	*f_username,
			Password:	"x",
			UID:		uint(*f_uid),
			GID:		uint(*f_gid),
			Gecos:		*f_comment,
			Dir:		*f_homedir,
			Shell:		*f_shell,
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("UID already exist in etcd:/etc/passwd/"))
		os.Exit(1)
	}

	resp, err = etcd_insert(
		path.Join("/etc/group/", strconv.Itoa(int(*f_uid))),
		Group{
			Groupname:	*f_username,
			Password:	"x",
			GID:		uint(*f_gid),
			Members:	[]string{ },
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("GID already exist in etcd:/etc/group/"))
		os.Exit(1)
	}

	
	resp, err = etcd_insert(
		path.Join("/etc/shadow/", strconv.Itoa(int(*f_uid))),
		Shadow{
			Username:			*f_username,
			Password:			hashedword,
			LastChange:			17920,
			MinChange:			0,
			MaxChange:			99999,
			PasswordWarn:		7,
			InactiveLockout:	-1,
			ExpirationDate:		-1,
			Reserved:			-1,
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("UID already exist in etcd:/etc/shadow/"))
		os.Exit(1)
	}

	fmt.Println("New user", *f_username, "added successfully.")
}

func user_delete(){
	// user delete --username="testuser"
	f_username := flag.String("username", "", "Username")
	flag.Parse()

	user := get_user(*f_username)
	if user.Username == "" {
		fmt.Println(errors.New("User not found in passwd"))
		os.Exit(1)
	}

	var err error
	_, err = etcd_delete(path.Join("/etc/passwd/", strconv.Itoa(int(user.UID))))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = etcd_delete(path.Join("/etc/group/", strconv.Itoa(int(user.GID))))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = etcd_delete(path.Join("/etc/shadow/", strconv.Itoa(int(user.UID))))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Deleted", *f_username ,"successfully.")
}

func group_add(){
	// group add --groupname="testguys" --gid=1501
	f_groupname := flag.String("groupname", "", "Groupname")
	f_gid := flag.Int("gid", getnextid(), "GID")
	flag.Parse()

	if group_exists(*f_groupname) {
		fmt.Println(errors.New("Group " + *f_groupname + " already exists."))
		os.Exit(1)
	}

	resp, err := etcd_insert(
		path.Join("/etc/group/", strconv.Itoa(int(*f_gid))),
		Group{
			Groupname:	*f_groupname,
			Password:	"x",
			GID:		uint(*f_gid),
			Members:	[]string{ },
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("GID already exist in etcd:/etc/group/"))
		os.Exit(1)
	}

	fmt.Println("New group", *f_groupname, "added!")
}

func group_delete(){
	// group delete --groupname="testguys"
	f_groupname := flag.String("groupname", "", "Groupname")
	flag.Parse()
	
	group := get_group(*f_groupname)
	if group.Groupname == "" {
		fmt.Println(errors.New("Group not found in passwd"))
		os.Exit(1)
	}

	_, err := etcd_delete(path.Join("/etc/group/", strconv.Itoa(int(group.GID))))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Deleted", *f_groupname ,"successfully.")
}

func group_add_member(){
	// group add-member --groupname="testguys" --username="testuser"
	f_username := flag.String("username", "", "Username")
	f_groupname := flag.String("groupname", "", "Groupname")
	flag.Parse()

	group := get_group(*f_groupname)
	if group.Groupname == "" {
		fmt.Println(errors.New("Group not found in passwd"))
		os.Exit(1)
	}

	if contains(group.Members, *f_username) {
		fmt.Println("User already in group.")
		return
	}
	
	group.Members = append(group.Members, *f_username)
	
	resp, err := etcd_update(
		path.Join("/etc/group/", strconv.Itoa(int(group.GID))),
		group,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("Failed to update entry in etcd:/etc/group/"))
		os.Exit(1)
	}
	
	fmt.Println("Updated", *f_groupname ,"successfully.")
}

func group_remove_member(){
	// group remove-member --groupname="testguys" --username="testuser"
	f_username := flag.String("username", "", "Username")
	f_groupname := flag.String("groupname", "", "Groupname")
	flag.Parse()

	group := get_group(*f_groupname)
	if group.Groupname == "" {
		fmt.Println(errors.New("Group not found in passwd"))
		os.Exit(1)
	}

	if !contains(group.Members, *f_username) {
		fmt.Println("User is not in the group.")
		return
	}

	var new_members []string
	for _, member := range group.Members {
		if member != *f_username {
			new_members = append(new_members, member)
		}
	}

	group.Members = new_members

	resp, err := etcd_update(
		path.Join("/etc/group/", strconv.Itoa(int(group.GID))),
		group,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !resp.Succeeded {
		fmt.Println(errors.New("Failed to update entry in etcd:/etc/group/"))
		os.Exit(1)
	}
	
	fmt.Println("Updated", *f_groupname ,"successfully.")
}
