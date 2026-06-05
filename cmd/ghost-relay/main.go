package main

import (
    "context"
    "flag"
    "log"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/multiformats/go-multiaddr"
)

var (
    listen = flag.String("listen", "/ip4/0.0.0.0/tcp/5001", "libp2p listen address")
    target = flag.String("target", "", "core multiaddr to forward to")
)

func main() {
    flag.Parse()
    if *target == "" {
        log.Fatal("need -target")
    }
    h, err := libp2p.New()
    if err != nil {
        log.Fatal(err)
    }
    targetAddr, _ := multiaddr.NewMultiaddr(*target)
    h.SetStreamHandler("/ghost/relay/1.0.0", func(s network.Stream) {
        coreStream, err := h.NewStream(context.Background(), s.Conn().RemotePeer(), "/ghost/core/1.0.0")
        if err != nil {
            log.Printf("relay stream error: %v", err)
            s.Reset()
            return
        }
        go copyStream(s, coreStream)
        go copyStream(coreStream, s)
    })
    log.Printf("Relay listening on %s", *listen)
    select {}
}

func copyStream(dst, src network.Stream) {
    defer dst.Close()
    defer src.Close()
    buf := make([]byte, 4096)
    for {
        n, err := src.Read(buf)
        if err != nil {
            return
        }
        if _, err := dst.Write(buf[:n]); err != nil {
            return
        }
    }
}
