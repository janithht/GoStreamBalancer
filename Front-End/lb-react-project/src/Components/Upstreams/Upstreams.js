import { useState } from 'react';
import './Upstreams.css';

function Upstreams() {
    const [selectedUpstream, setSelectedUpstream] = useState('');
    const upstreams = ['Upstream 1', 'Upstream 2', 'Upstream 3'];

    const requestData = {
        'Upstream 1': { success: 120, error: 5, warning: 10 },
        'Upstream 2': { success: 200, error: 15, warning: 20 },
        'Upstream 3': { success: 80, error: 2, warning: 6 }
    };

    const serverStatus = {
        'Upstream 1': ['Server 1: Healthy', 'Server 2: Down'],
        'Upstream 2': ['Server 1: Healthy', 'Server 2: Healthy'],
        'Upstream 3': ['Server 1: Down', 'Server 2: Healthy']
    };

    const handleSelectUpstream = (event) => {
        setSelectedUpstream(event.target.value);
    };

    return (
        <section id="upstreams">
            <div className="upstreams">
                <h2>Upstreams</h2>
                <select onChange={handleSelectUpstream} value={selectedUpstream}>
                    <option value="">Select an Upstream</option>
                    {upstreams.map(upstream => (
                        <option key={upstream} value={upstream}>{upstream}</option>
                    ))}
                </select>
                <div className="dataBoxes">
                    <div className="requestsBox">
                        <h3>Total Requests</h3>
                        {selectedUpstream && (
                            <ul>
                                <li>Success: {requestData[selectedUpstream]?.success || 0}</li>
                                <li>Error: {requestData[selectedUpstream]?.error || 0}</li>
                                <li>Warning: {requestData[selectedUpstream]?.warning || 0}</li>
                            </ul>
                        )}
                    </div>
                    <div className="statusBox">
                        <h3>Server Statuses</h3>
                        {selectedUpstream && (
                            <ul>
                                {serverStatus[selectedUpstream].map(status => (
                                    <li key={status}>{status}</li>
                                ))}
                            </ul>
                        )}
                    </div>
                </div>
            </div>
        </section>
    );
}

export default Upstreams;
