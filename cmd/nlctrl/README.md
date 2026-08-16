# nlctrl

nlctrl is a utility to get information about the registered Generic Netlink families on the system.


## installation

```sh
go install go.jamescun.com/netlink/cmd/nlctrl@latest
```

Also see [releases](https://github.com/jamescun/netlink/releases) for pre-built binaries and packages.


## usage

Display usage information:

```sh
$ nlctrl --help
nlctrl
get generic netlink family information

USAGE: nlctrl [options...] [family name]

FAMILY NAME:
  When a family name is specified, only information about that family is
  returned. Otherwise all families registered on the system will be returned.

OPTIONS:
  -h --help  show usage information
```

Retrieve information about a single family:

```sh
$ nlctrl wireguard
Family:   30
Name:     wireguard v1
Desc:     Netlink protocol to control WireGuard network devices.

Commands:
  WG_CMD_GET_DEVICE
    Flags:  CMD_CAP_DUMP CMD_CAP_HASPOL UNS_ADMIN_PERM
  WG_CMD_SET_DEVICE
    Flags:  CMD_CAP_DO CMD_CAP_HASPOL UNS_ADMIN_PERM
```

Retrieve information about all families:

```sh
$ nlctrl
Family:   16
Name:     nlctrl v2
Desc:     genetlink meta-family that exposes information about all genetlink families registered in the kernel (including itself).

Commands:
  CTRL_CMD_GETFAMILY
    Flags:  CMD_CAP_DO CMD_CAP_DUMP CMD_CAP_HASPOL
  CTRL_CMD_GETPOLICY
    Flags:  CMD_CAP_DUMP CMD_CAP_HASPOL

Multicast Groups:
    16: notify

...
```
