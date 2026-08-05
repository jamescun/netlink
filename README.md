# netlink

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink.svg)](https://pkg.go.dev/go.jamescun.com/netlink)

This package contains implements the Linux Kernel Netlink protocol, for interacting with the systems network stack and related subsystems.


## core packages

### netlink

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink.svg)](https://pkg.go.dev/go.jamescun.com/netlink)

```sh
go get -u go.jamescun.com/netlink
```

The `netlink` package implements the encoding of and exchanging Netlink messages with a unix socket.


### genetlink

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/genetlink.svg)](https://pkg.go.dev/go.jamescun.com/netlink/genetlink)

```sh
go get -u go.jamescun.com/netlink/genetlink
```

The `genetlink` package contains helpers for interacting with the [Generic Netlink](https://wiki.linuxfoundation.org/networking/generic_netlink_howto) protocol and it's associated families.

It also implements a client for the Generic Netlink [Controller](https://pkg.go.dev/go.jamescun.com/netlink/genetlink#Controller), for discovering what families are available and how to connect to them.


## additional packages

### rtnetlink

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/rtnetlink.svg)](https://pkg.go.dev/go.jamescun.com/netlink/rtnetlink)

```sh
go get -u go.jamescun.com/netlink/rtnetlink
```


### wireguard

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/wireguard.svg)](https://pkg.go.dev/go.jamescun.com/netlink/wireguard)

```sh
go get -u go.jamescun.com/netlink/wireguard
```
