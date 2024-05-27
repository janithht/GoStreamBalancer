import { useState } from 'react';
import './Header.css';
import logo from '../../Assets/lblogo.png';

function Header() {
    const [isHttp, setIsHttp] = useState(true);

    const toggleLoadBalancerType = () => {
        setIsHttp(!isHttp);
    };

    return (
        <section id="header">
            <header className="header">
            <img src={logo} alt="Logo" className="logo" /> 
                <h1 className="loadBalancerName"><i>Go Stream Balancer</i></h1>
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
