import { useState } from 'react';
import './Header.css';

function Header() {
    const [isHttp, setIsHttp] = useState(true);

    const toggleLoadBalancerType = () => {
        setIsHttp(!isHttp);
    };

    return (
        <section id="header">
            <header className="header">
                <h1 className="loadBalancerName">My Load Balancer</h1>
                <div className="switches">
                    <button
                        className={`switch ${isHttp ? "active" : ""}`}
                        onClick={toggleLoadBalancerType}
                    >
                        HTTP
                    </button>
                    <button
                        className={`switch ${!isHttp ? "active" : ""}`}
                        onClick={toggleLoadBalancerType}
                    >
                        TCP
                    </button>
                </div>
            </header>
        </section>
    );
}

export default Header;
