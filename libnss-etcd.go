package main

import (
	. "github.com/protosam/go-libnss"
	. "github.com/protosam/go-libnss/structs"
)

// Placeholder main() stub is neccessary for compile.
func main() {}

func init(){
	// We set our implementation to "LibNssEtcd", so that go-libnss will use the methods we create
	SetImpl(LibNssEtcd{})
}

// We're creating a struct that implements LIBNSS stub methods.
type LibNssEtcd struct { LIBNSS }

////////////////////////////////////////////////////////////////
// Passwd Methods
////////////////////////////////////////////////////////////////

// PasswdAll() will populate all entries for libnss
func (self LibNssEtcd) PasswdAll() (Status, []Passwd) {
	if !etcd_client_success {
		return StatusUnavail, []Passwd{}
	}
	if len(PasswdDB) == 0 {
		return StatusUnavail, []Passwd{}
	}
	return StatusSuccess, PasswdDB
}

// PasswdByName() returns a single entry by name.
func (self LibNssEtcd) PasswdByName(name string) (Status, Passwd) {
	if !etcd_client_success {
		return StatusUnavail, Passwd{}
	}

	if len(PasswdDB) == 0 {
		return StatusUnavail, Passwd{}
	}
	for _, entry := range PasswdDB {
		if entry.Username == name {
			return StatusSuccess, entry
		}
	}

	return StatusNotfound, Passwd{}
}

// PasswdByUid() returns a single entry by uid.
func (self LibNssEtcd) PasswdByUid(uid uint) (Status, Passwd) {
	if !etcd_client_success {
		return StatusUnavail, Passwd{}
	}

	if len(PasswdDB) == 0 {
		return StatusUnavail, Passwd{}
	}
	for _, entry := range PasswdDB {
		if entry.UID == uid {
			return StatusSuccess, entry
		}
	}

	return StatusNotfound, Passwd{}
}


////////////////////////////////////////////////////////////////
// Group Methods
////////////////////////////////////////////////////////////////
// endgrent
func (self LibNssEtcd) GroupAll() (Status, []Group) {
	if !etcd_client_success {
		return StatusUnavail, []Group{}
	}

	if len(GroupDB) == 0 {
		return StatusUnavail, []Group{}
	}

	return StatusSuccess, GroupDB
}

// getgrent
func (self LibNssEtcd) GroupByName(name string) (Status, Group) {
	if !etcd_client_success {
		return StatusUnavail, Group{}
	}

	if len(GroupDB) == 0 {
		return StatusUnavail, Group{}
	}
	for _, entry := range GroupDB {
		if entry.Groupname == name {
			return StatusSuccess, entry
		}
	}

	return StatusNotfound, Group{}
}

// getgrnam
func (self LibNssEtcd) GroupByGid(gid uint) (Status, Group) {
	if !etcd_client_success {
		return StatusUnavail, Group{}
	}

	if len(GroupDB) == 0 {
		return StatusUnavail, Group{}
	}
	for _, entry := range GroupDB {
		if entry.GID == gid {
			return StatusSuccess, entry
		}
	}

	return StatusNotfound, Group{}
}

////////////////////////////////////////////////////////////////
// Shadow Methods
////////////////////////////////////////////////////////////////
// endspent
func (self LibNssEtcd) ShadowAll() (Status, []Shadow) {
	if !etcd_client_success {
		return StatusUnavail, []Shadow{}
	}

	if len(ShadowDB) == 0 {
		return StatusUnavail, []Shadow{}
	}

	return StatusSuccess, ShadowDB
}

// getspent
func (self LibNssEtcd) ShadowByName(name string) (Status, Shadow) {
	if !etcd_client_success {
		return StatusUnavail, Shadow{}
	}

	if len(ShadowDB) == 0 {
		return StatusUnavail, Shadow{}
	}
	for _, entry := range ShadowDB {
		if entry.Username == name {
			return StatusSuccess, entry
		}
	}

	return StatusNotfound, Shadow{}
}
