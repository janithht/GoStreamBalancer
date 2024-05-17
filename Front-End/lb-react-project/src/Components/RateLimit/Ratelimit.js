import './Ratelimit.css';

function RateLimiting() {
    // Sample data for demonstration purposes
    const rateLimits = [
        { upstream: "Upstream 1", limit: 100, current: 95 },
        { upstream: "Upstream 2", limit: 200, current: 202 },
        { upstream: "Upstream 3", limit: 50, current: 45 }
    ];

    return (
        <section id="rateLimiting">
            <div className="rateLimiting">
                <h2>Rate Limiting</h2>
                <div className="limitsBox">
                    {rateLimits.map((limit, index) => (
                        <div key={index} className="limitDetails">
                            <h3>{limit.upstream}</h3>
                            <p>Limit: {limit.limit}</p>
                            <p>Current: {limit.current}</p>
                            {limit.current > limit.limit && <p className="overLimit">Limit Exceeded!</p>}
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

export default RateLimiting;
