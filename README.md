# GhostRelay

Decentralized peer-to-peer command and control framework. Uses libp2p for network communication, IPFS for payload hosting, and process hollowing for stealth execution. No central servers, no fixed IP addresses.

## Features

- Peer-to-peer communication via libp2p (Kademlia DHT, Noise encryption)
- Payload delivery through IPFS network (fileless download)
- AES-256-GCM encryption with dynamic nonce per message
- Automatic agent reconnection after core restart
- Process hollowing injection into legitimate Windows processes (explorer.exe)
- Web-based operator interface (React)
- Persistent storage using BoltDB (commands, results, agent tracking)
- Optional libp2p relays for traffic obfuscation

## Architecture

Core (controller) accepts connections from agents and console. Agents connect to core through optional relays. All traffic is encrypted. Payloads are hosted on IPFS and pulled by agents on demand.

## Requirements

- Go 1.22+
- Rust 1.75+
- Node.js 18+
- Make
- IPFS Kubo (optional, for local payload hosting)

## Build

```bash
git clone https://github.com/yourusername/ghost-relay.git
cd ghost-relay
make all
make windows-agent
cd web-ui && npm install && npm run build
