# eBPF XDP Redirect Demo

This is a simple example of doing packet redirection in [XDP][xdp] programs
using [ebpf-go][ebpf-go].

[xdp]:https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/
[ebpf-go]:https://github.com/cilium/ebpf

## Requirements

* A [Linux][linux] system with kernel version `5.19`+
* [Docker][docker] version `20.10`+
* [Netcat][nc] version `0.7.1`+

> **Note**: This demo assumes a standard [Docker][docker] configuration where
> a `docker0` bridge device exists, the default network's subnet is
> `172.17.0.0/24`, and that you have no other containers running (so that our
> test container gets the address `172.17.0.2`). If this is not the case, make
> manual adjustments to the IPs and ethernet addresses in `main.go` according
> to your environment.

[linux]:https://kernel.org/
[docker]:https://www.docker.com/
[nc]:https://man.archlinux.org/man/netcat.1.en

## Running

This example demonstrate redirecting [UDP][udp] packets on port `9875` from the
`lo` (loopback) interface on your host system to a container on that system.

The container can be set up with:

```console
$ ./env.sh
```

The container will run `nc -kul 172.17.0.2 9875` and will listen for our UDP
packets.

In another terminal, watch the output of the container:

```console
$ docker logs -f udp-listener
```

> **Note**: Prior to attaching the XDP program, you can test to verify that the
> UDP listen server is working by running `echo "test" | nc -u 172.17.0.2 9875`,
> you should see "test" show up in the output of `docker logs`.

Build and run with:

```console
$ make && sudo ./demo lo docker0
2022/10/21 16:24:32 Attached XDP program to iface "lo" (index 1) and iface "docker0" (index 2)
2022/10/21 16:24:32 Press Ctrl-C to exit and remove the program
```

> **Note**: You can optionally view the trace output of the XDP program by
> running the following in another terminal:
> `sudo cat /sys/kernel/debug/tracing/trace_pipe`

> **Note**: You can keep a count of the number of times your program redirects
> a packet by running:
> `sudo bpftrace -e 'tracepoint:xdp:xdp_redirect { @cnt[probe] = count(); }'`

> **Note**: You can dump the raw packets being processed by the XDP program
> (similar to `tcpdump`) by running these:
>  - `xdpdump -i lo -x --rx-capture entry,exit`
>  - `xdpdump -i docker0 -x --rx-capture entry,exit`

Now send some data:

```console
$ echo "test" | nc -u 127.0.0.1 9875
```

If everything worked properly, you should see `test` in the output of
`docker logs` for the container.

> **Note**: Currently this attaches a "placeholder" XDP program to `docker0`
> as is [required][0] for XDP redirects to work. Alternatively you can turn
> on the Generic Receive Offload (GRO) feature for your destination interface
> instead of attaching a placeholder XDP program if your kernel is compatible.
> Enable it with: `ethtool -K docker0 gro on`.

[udp]:https://www.cloudflare.com/learning/ddos/glossary/user-datagram-protocol-udp/
[0]:https://github.com/torvalds/linux/blob/9e9fb7655ed585da8f468e29221f0ba194a5f613/samples/bpf/xdp_redirect.bpf.c#L42

# License

The contents of this demo are licensed under the terms of the [General Public
License, v2][gpl] or [MIT License][mit] at your option.

[gpl]:https://github.com/shaneutt/ebpf-xdp-golang-redirect-demo/blob/main/LICENSE-GPL
[mit]:https://github.com/shaneutt/ebpf-xdp-golang-redirect-demo/blob/main/LICENSE-MIT
