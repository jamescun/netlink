### tuntap

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/tuntap.svg)](https://pkg.go.dev/go.jamescun.com/netlink/tuntap)

```sh
go get -u go.jamescun.com/netlink/tuntap
```

The `tuntap` package is an implementation of [Linux TUNTAP](https://www.kernel.org/doc/html/latest/networking/tuntap.html) devices for creating and using Layer-3 TUN and Layer-3 TAP devices.

While TUNTAP devices do not directly make use of netlink, it is the only device type that isn't created and configured through rtnetlink. This package bridges that gap, and is built in such a way as to be fully compatible with the rtnetlink package.


## Examples

Below are some examples of how to use the `tuntap` package.

> [!NOTE]
> Error handling is omitted in these examples for brevity.


### Creating and configuring a device

```go
package main

import (
	"net/netip"

	"go.jamescun.com/netlink/rtnetlink"
	"go.jamescun.com/netlink/rtnetlink/addr"
	"go.jamescun.com/netlink/rtnetlink/link"
	"go.jamescun.com/netlink/tuntap"
)

func main() {
	rt, _ := rtnetlink.New()

	// create a Layer-3 TUN device, without the intermediate TUNTAP packet
	// information header.
	tun, _ := tuntap.NewTUN("foo0", tuntap.NoPacketInfo())

	// configure the TUN device with the `100.64.0.1/24` IP Address.
	rt.AddAddr(
		tun.Index(),
		netip.MustParsePrefix("100.64.0.1/24"),
		addr.FlagPermanent,
	)

	// bring the TUN device up to receive packets.
	rt.SetLink(tun.Index(), link.Up)

	// use tun.Read() and tun.Write() to receive and transmit packets.
}
```
