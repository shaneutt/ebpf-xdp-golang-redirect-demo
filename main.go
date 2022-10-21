package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf xdp.c -- -I../headers

type backend struct {
	saddr   uint32
	daddr   uint32
	hwaddr  [6]uint8
	ifindex uint16
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Please specify a main and destination network interface")
	}

	ifaceName := os.Args[1]
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("lookup network iface %q: %s", ifaceName, err)
	}
	ifaceDestName := os.Args[2]
	ifaceDest, err := net.InterfaceByName(ifaceDestName)
	if err != nil {
		log.Fatalf("lookup network iface %q: %s", ifaceDestName, err)
	}

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %s", err)
	}
	defer objs.Close()

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProgFunc,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatalf("could not attach XDP program: %s", err)
	}
	defer l.Close()

	l2, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.BpfRedirectPlaceholder,
		Interface: ifaceDest.Index,
	})
	if err != nil {
		log.Fatalf("could not attach XDP program: %s", err)
	}
	defer l2.Close()

	log.Printf("Attached XDP program to iface %q (index %d) and iface %q (index %d)", iface.Name, iface.Index, ifaceDest.Name, ifaceDest.Index)
	log.Printf("Press Ctrl-C to exit and remove the program")

	b := backend{
		saddr:   ip2int("172.17.0.1"),
		daddr:   ip2int("172.17.0.2"),
		hwaddr:  hwaddr2bytes("02:42:ac:11:00:02"),
		ifindex: uint16(ifaceDest.Index),
	}

	if err := objs.Backends.Update(ip2int("127.0.0.1"), b, ebpf.UpdateAny); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for {
	}
}

func ip2int(ip string) uint32 {
	ipaddr := net.ParseIP(ip)
	return binary.LittleEndian.Uint32(ipaddr.To4())
}

func hwaddr2bytes(hwaddr string) [6]byte {
	parts := strings.Split(hwaddr, ":")
	if len(parts) != 6 {
		panic("invalid hwaddr")
	}

	var hwaddrB [6]byte
	for i, hexPart := range parts {
		bs, err := hex.DecodeString(hexPart)
		if err != nil {
			panic(err)
		}
		if len(bs) != 1 {
			panic("invalid hwaddr part")
		}
		hwaddrB[i] = bs[0]
	}

	return hwaddrB
}
