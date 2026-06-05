import React, { useState, useEffect } from 'react';
import axios from 'axios';

function App() {
    const [agents, setAgents] = useState([]);
    const [selectedAgent, setSelectedAgent] = useState('');
    const [command, setCommand] = useState('');
    const [output, setOutput] = useState('');

    const fetchAgents = async () => {
        const res = await axios.get('/api/agents');
        setAgents(res.data);
    };

    const sendCommand = async () => {
        if (!selectedAgent || !command) return;
        const res = await axios.post('/api/command', { agent: selectedAgent, cmd: command });
        setOutput(res.data.output);
    };

    useEffect(() => {
        fetchAgents();
        const interval = setInterval(fetchAgents, 3000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div style={{ padding: '20px', fontFamily: 'monospace' }}>
            <h1>GhostRelay Web UI</h1>
            <h2>Agents</h2>
            <ul>
                {agents.map(a => <li key={a.id}>{a.name} ({a.id})</li>)}
            </ul>
            <select value={selectedAgent} onChange={e => setSelectedAgent(e.target.value)}>
                <option value="">Select agent</option>
                {agents.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
            </select>
            <input value={command} onChange={e => setCommand(e.target.value)} placeholder="command" />
            <button onClick={sendCommand}>Execute</button>
            <pre>{output}</pre>
        </div>
    );
}

export default App;
