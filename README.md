# netlink

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink.svg)](https://pkg.go.dev/go.jamescun.com/netlink)

This package contains implements the Linux Kernel Netlink protocol, for interacting with the systems network stack and related subsystems.

> [!WARNING]
> This package is still **under heavy development**, it's API may change unprompted, or may behave in unexpected or destructive ways.
>
> **USE AT YOUR OWN RISK!**

Guiding Principles:

* Ecosystem compatibility:
  * Use interfaces defined in the Go standard library, such as [encoding.BinaryMarshaler](https://pkg.go.dev/encoding#BinaryMarshaler) and [encoding.BinaryUnmarshaler](https://pkg.go.dev/encoding#BinaryUnmarshaler).
* No third-party dependencies or shell commands.

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

The `rtnetlink` package is an implementation of the `rtnetlink` family, for interacting with system network interfaces, addresses and routing and network neighbors.

The client is implemented in [rtnetlink](https://pkg.go.dev/go.jamescun.com/netlink/rtnetlink), while the individual families are implemented in:

* [netlink/rtaddr](https://pkg.go.dev/go.jamescun.com/netlink/rtnetlink/rtaddr)
* [netlink/rtlink](https://pkg.go.dev/go.jamescun.com/netlink/rtnetlink/rtlink)
* [netlink/rtroute](https://pkg.go.dev/go.jamescun.com/netlink/rtnetlink/rtroute)

See usage examples in the [README.md](rtnetlink/README.md).


### tuntap

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/tuntap.svg)](https://pkg.go.dev/go.jamescun.com/netlink/tuntap)

```sh
go get -u go.jamescun.com/netlink/tuntap
```

The `tuntap` package is an implementation of [Linux TUNTAP](https://www.kernel.org/doc/html/latest/networking/tuntap.html) devices for creating and using Layer-3 TUN and Layer-3 TAP devices.

See usage examples in the [README.md](tuntap/README.md).


### wireguard

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/wireguard.svg)](https://pkg.go.dev/go.jamescun.com/netlink/wireguard)

```sh
go get -u go.jamescun.com/netlink/wireguard
```
