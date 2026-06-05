package main

import (
    "bufio"
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "ghost-relay/internal/crypto"
    "ghost-relay/internal/protobuf"

    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
    "google.golang.org/protobuf/proto"
)

var (
    coreAddr = flag.String("core", "", "multiaddr of ghost core (e.g., /ip4/127.0.0.1/tcp/4001/p2p/...)")
    keyHex   = flag.String("key", "", "encryption key (same as core)")
)

type model struct {
    ready    bool
    viewport viewport.Model
    input    string
    messages []string
    corePeer peer.ID
    crypto   *crypto.Crypto
    stream   *libp2p.Stream
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            if m.input != "" {
                m.messages = append(m.messages, "> "+m.input)
                
                cmd := &protobuf.Command{Id: time.Now().String(), Type: "exec", Args: m.input}
                data, _ := proto.Marshal(cmd)
                m.stream.Write(data)
                m.input = ""
            }
        case "backspace":
            if len(m.input) > 0 {
                m.input = m.input[:len(m.input)-1]
            }
        default:
            m.input += msg.String()
        }
        m.viewport.SetContent(strings.Join(m.messages, "\n"))
        m.viewport.GotoBottom()
    }
    return m, nil
}

func (m model) View() string {
    header := lipgloss.NewStyle().Background(lipgloss.Color("#000")).Foreground(lipgloss.Color("#0f0")).Padding(1).Render("Ghost Console – type commands (exit to quit)")
    inputLine := "> " + m.input
    return header + "\n" + m.viewport.View() + "\n" + inputLine
}

func main() {
    flag.Parse()
    if *coreAddr == "" || *keyHex == "" {
        log.Fatal("need -core and -key")
    }
    crypt, _ := crypto.NewCrypto(*keyHex)
    addr, _ := multiaddr.NewMultiaddr(*coreAddr)
    pi, _ := peer.AddrInfoFromP2pAddr(addr)
    host, _ := libp2p.New()
    stream, err := host.NewStream(context.Background(), pi.ID, "/ghost/console/1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    
    stream.Write(crypt.KeyHash())
    buf := make([]byte, 2)
    stream.Read(buf)
    if string(buf) != "OK" {
        log.Fatal("auth failed")
    }

    p := tea.NewProgram(model{
        viewport: viewport.New(80, 20),
        crypto:   crypt,
        stream:   stream,
        corePeer: pi.ID,
    })
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
    stream.Close()
}
