### rt

[![Go Reference](https://pkg.go.dev/badge/go.jamescun.com/netlink/rt.svg)](https://pkg.go.dev/go.jamescun.com/netlink/rt)

```sh
go get -u go.jamescun.com/netlink/rtnetlink
```

The `rt` package is an implementation of the rtnetlink family, for interacting with system network interfaces, addresses and routing and network neighbors.

The client is implemented in [rtnetlink](https://pkg.go.dev/go.jamescun.com/netlink/rt), while the individual families are implemented in:

* [netlink/rt/rtaddr](https://pkg.go.dev/go.jamescun.com/netlink/rt/rtaddr)
* [netlink/rt/rtlink](https://pkg.go.dev/go.jamescun.com/netlink/rt/rtlink)
* [netlink/rt/rtroute](https://pkg.go.dev/go.jamescun.com/netlink/rt/rtroute)


## Examples

Below are some examples of how to use the `rt` client.

> [!NOTE]
> Error handling is omitted in these examples for brevity.


### List links

```go
package main

import (
	"fmt"

	"go.jamescun.com/netlink/rt/rtlink"
)

func main() {
	client, _ := rtlink.New()

	// get a list of links for all families (including IPv4 and IPv6),
	// filtering expensive attributes such as Virtual Functions and Statistics.
	links, _ := client.ListLinks(rtlink.ALL, rtlink.VF|rtlink.SKIP_STATS)

	for _, link := range links {
		fmt.Printf("Link: %d, Name: %s\n", link.Index, link.Name)
	}
}
```


### Creating a link

```go
package main

import (
	"go.jamescun.com/netlink/rt/rtlink"
)

func main() {
	client, _ := rtlink.New()

	// create a new Wireguard link called `wg0` with an MTU of 1420, and bring
	// the interface up.
	client.CreateLink(
		"wg0",
		rtlink.Generic("wireguard"),
		rtlink.MTU(1420),
		rtlink.Up,
	)
}
```


### Configure a link

```go
package main

import (
	"go.jamescun.com/netlink/rt/rtlink"
)

func main() {
	client, _ := rtlink.New()

	// configure the `wg0` link to up.
	client.ConfigureLink("wg0", rtlink.Up)
}
```
