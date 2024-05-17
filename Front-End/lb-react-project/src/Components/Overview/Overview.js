import './Overview.css';

function Overview() {
    // Sample data for demonstration purposes
    const health = "Healthy";
    const ports = ["80", "443", "8080"];

    return (
        <section id="overview">
            <div className="overview">
                <div className="healthBox">
                    <h2>Load Balancer Health</h2>
                    <p>Status: {health}</p>
                </div>
                <div className="entryPoints">
                    <h2>Entry Points</h2>
                    <ul>
                        {ports.map(port => (
                            <li key={port}>Port {port}</li>
                        ))}
                    </ul>
                </div>
            </div>
        </section>
    );
}

export default Overview;