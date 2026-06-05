package libp2p

import (
    "context"
    "fmt"
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

func NewHost(ctx context.Context, listenPort int, bootstrapPeers []string) (host.Host, *dht.IpfsDHT, error) {
    listenAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
    h, err := libp2p.New(libp2p.ListenAddrs(listenAddr))
    if err != nil {
        return nil, nil, err
    }
    kademliaDHT, err := dht.New(ctx, h)
    if err != nil {
        return nil, nil, err
    }
    for _, bp := range bootstrapPeers {
        addr, _ := multiaddr.NewMultiaddr(bp)
        pi, _ := peer.AddrInfoFromP2pAddr(addr)
        if pi != nil {
            go h.Connect(ctx, *pi)
        }
    }
    return h, kademliaDHT, nil
}
