# go-libnss-etcd
A libnss module and commands for managing additional users in `etcd`.

Note: At this time go-libnss-etcd works. The `nss-etcd-passwd` needs rigorous testing before it should ever go into production, because it is expected to always run as root. If you need that command, I recommend not setting the sticky bit for it to run as root in production, so that unprivileged users can run it. 

# Simple Installation (Quick and Lazy)
Do the steps in the section `Locking Down ETCD`  
Do the steps in the section `Configuration Files`  
Run this as root:
```
# make && make install
```
Do the showing the example of configuring `/etc/nsswitch.conf` mentioned in the `Installing libnss_etcd.so.2` section.

After that you should be done. Just use the management tools documented in the last section of this README.

# Building Manually
There are 3 parts to compile this. There is the `libnss_etcd.so.2` shared library, the `nss-etcd-manage` CLI tool, and the `nss-etcd-passwd` CLI tool.

Compiling `libnss_etcd.so.2` is done by running the following command.
```
CGO_CFLAGS="-g -O2 -D __LIB_NSS_NAME=etcd" go build --buildmode=c-shared -o libnss_etcd.so.2 libnss-etcd.go etcd-db.go
```

The `nss-etcd-manage` CLI tool is used for managing users and groups in etcd. Compiling it is done by running:
```
go build -o nss-etcd-manage etcd-db.go nss-etcd-manage.go
```

Lastly is `nss-etc-passwd` for user password changes. It is compiled by running
```
go build -o nss-etcd-passwd etcd-db.go nss-etcd-passwd.go
```

# Installation
The install process is broken up into a few parts. You need to lock down your `etcd` keystore a bit, putting compiled files into place, and setting the sticky bit on `nss-etc-passwd`. Just do the following after compiling the binaries and you should be in business.

Additional note: `selinux` does not like allowing these users to login through SSH. I personally just disable selinux, because it causes a lot of unneccessary hassle. However if you can't disable `selinux` because you need the additional security it provides, please do troubleshoot this and share with me in the issues section what steps I need to add for setting contexts and whatnot. I'll be happy to add them to this readme.

## Locking Down ETCD
You will need to set password for root, enable authentication, create a read-only and read-write role for `go-libnss-etcd`. Below are some dead simple copy/pasta you can use. The variables you're exporting should probably be added to your `.bashrc` file for easy use later.

```
# export ETCDCTL_API=3
# etcdctl user add 'root:YOUR_ROOT_ETCD_PASSWORD_HERE'
# export ETCDCTL_USER='root:YOUR_ROOT_ETCD_PASSWORD_HERE'
# etcdctl user add nss-ro:YOUR_READ_ONLY_PASSWORD_HERE
# etcdctl user add nss-rw:YOUR_READ_WRITE_PASSWORD_HERE
# etcdctl role add nss-ro
# etcdctl role add nss-rw
# etcdctl role grant-permission nss-ro --prefix=true read /etc/passwd
# etcdctl role grant-permission nss-ro --prefix=true read /etc/group
# etcdctl role grant-permission nss-rw --prefix=true readwrite /etc/passwd
# etcdctl role grant-permission nss-rw --prefix=true readwrite /etc/group
# etcdctl role grant-permission nss-rw --prefix=true readwrite /etc/shadow
# etcdctl user grant-role nss-ro nss-ro
# etcdctl user grant-role nss-rw nss-rw
```

## Configuration Files
For non-privileged users `libnss-etcd.so.2` and `nss-etcd-passwd` will use `/etc/nss-etcd.conf`. The contents of this config file is in JSON format with the following data:
```
{
	"Endpoints": ["http://localhost:2379"],
	"DialTimeout": 2,
	"Username": "nss-ro",
	"Password":	"YOUR_READ_ONLY_PASSWORD_HERE",
	"MinXID": 2000
}
```

For privileged users like root, `libnss-etcd.so.2`, `nss-etcd-manage`, and `nss-etcd-passwd` will use `/etc/nss-etcd-root.conf`. The contents of this config file is in JSON format with the following data:
```
{
	"Endpoints": ["http://localhost:2379"],
	"DialTimeout": 2,
	"Username": "nss-rw",
	"Password":	"YOUR_READ_WRITE_PASSWORD_HERE",
	"MinXID": 2000
}
```

After you've created the config files, the next step is to set permissions on the config files with appropriate read/write access:
```
# chmod u=rw,g=r,o=r /etc/nss-etcd.conf
# chmod u=rw,g=,o= /etc/nss-etcd-root.conf
```

## Installing Binaries
Copy `nss-etcd-manage` to `/sbin/` and `nss-etcd-passwd` to `/bin`.
You will need to set the sticky bit for `nss-etcd-passwd` to run as root so it can update the shadow entries when users want to update their passwords:
```
cp nss-etcd-manage /sbin/
cp nss-etcd-passwd /bin/
chown root:root /bin/nss-etcd-passwd /sbin/nss-etcd-manage
chmod u=rwx,g=rx,o=rx /bin/nss-etcd-manage
chmod u=rwxs,g=rx,o=rx /bin/nss-etcd-passwd
```

## Installing libnss_etcd.so.2
Copy `libnss_etcd.so.2` to `/lib64/` or where ever your shared library directory is.
Update `/etc/nsswitch.conf` to contain `etcd` like so:
```
passwd:     files etcd sss
shadow:     files etcd sss
group:      files etcd sss
```
And now the users stored in `etcd` should be visible to the system.

# Management Tools
`nss-etcd-manage` is used to manage user and group entries. 

To add new users to libnss-etcd:
```
# nss-etcd-manage user add --username="testuser" --password="password" --uid=2000 --gid=2000 --comment="Is stored in etcd." --homedir="/home/testuser" --shell="/bin/bash"
```

To delete existing users from libnss-etcd
```
# nss-etcd-manage user delete --username="testuser"
```

To add a new group to libnss-etcd:
```
# nss-etcd-manage group add --groupname="testguys" --gid=2001
```

To delete a group from libnss-etcd:
```
# nss-etcd-manage group delete --groupname="testguys"
```

To add a member to a group in libnss-etcd:
```
# nss-etcd-manage group add-member --groupname="testguys" --username="testuser"
```

To remove a member from a group in libnss-etcd:
```
# nss-etcd-manage group remove-member --groupname="testguys" --username="testuser"
```

`nss-etcd-passwd` is used to change passwords for users.

As the root user you can run the following to change a user's password. Only root can user the --username flag.
```
# nss-etcd-passwd --username john --password new_password
```

As a libnss-etcd user, you can run the following to change your own password.
```
# nss-etcd-passwd --password new_password
```
