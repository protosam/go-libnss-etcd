MACHINE := $(shell uname -m)

ifeq ($(MACHINE), x86_64)
libdir = /usr/lib64/
endif
ifeq ($(MACHINE), i686)
libdir = /usr/lib/
endif

build:
	CGO_CFLAGS="-g -O2 -D __LIB_NSS_NAME=etcd" go build --buildmode=c-shared -o libnss_etcd.so.2 libnss-etcd.go etcd-db.go
	go build -o nss-etcd-manage etcd-db.go nss-etcd-manage.go
	go build -o nss-etcd-passwd etcd-db.go nss-etcd-passwd.go

install:
	/bin/cp -fv libnss_etcd.so.2 $(libdir)
	/bin/cp -fv nss-etcd-manage /sbin/
	/bin/cp -fv nss-etcd-passwd /bin/
	/bin/chown root:root /bin/nss-etcd-passwd /sbin/nss-etcd-manage
	/bin/chmod u=rwx,g=rx,o=rx /bin/nss-etcd-manage
	/bin/chmod u=rwxs,g=rx,o=rx /bin/nss-etcd-passwd

uninstall:
	rm -rf /sbin/libnss_etcd.so.2 /sbin/nss-etcd-manage /bin/nss-etcd-passwd

clean:
	rm -rf libnss_etcd.so.2 nss-etcd-manage nss-etcd-passwd
