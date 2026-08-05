package main

import (
	"fmt"

	"go.jamescun.com/netlink/rtnetlink"

	"golang.org/x/sys/unix"
)

func main() {
	client, err := rtnetlink.New()
	if err != nil {
		panic(err)
	}

	links, err := client.ListLinks(unix.AF_UNSPEC)
	if err != nil {
		panic(err)
	}

	fmt.Println("INDEX   NAME               MAC")
	fmt.Println("-----   ----               ---")

	for _, link := range links {
		attrs := link.LinkAttrs()

		fmt.Printf("%-5d   %-16s   %s\n", attrs.Index, attrs.Name, attrs.Address)
	}
}
