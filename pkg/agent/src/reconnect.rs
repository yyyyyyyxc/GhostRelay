use libp2p::swarm::NetworkBehaviour;
use libp2p::ping;

#[derive(NetworkBehaviour)]
pub struct ReconnectBehaviour {
    ping: ping::Behaviour,
}

impl Default for ReconnectBehaviour {
    fn default() -> Self {
        Self { ping: ping::Behaviour::default() }
    }
}
