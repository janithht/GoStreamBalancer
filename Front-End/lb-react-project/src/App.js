import './App.css';
import Header from './Components/Header/Header';
import Overview from './Components/Overview/Overview';
import Upstreams from './Components/Upstreams/Upstreams';
import Footer from './Components/Footer/Footer';
import Metrics from './Components/Metrics/Metrics';
import Metricstcp from './Components/Metricstcp/Metricstcp';
import { useState } from 'react';

function App() {
  const [isHttp, setIsHttp] = useState(true);

  const toggleLoadBalancerType = () => {
    setIsHttp(!isHttp);
  };

  return (
    <div className="App">
      <Header className="App-header" isHttp={isHttp} toggleLoadBalancerType={toggleLoadBalancerType} />
      <main className="App-main">
        {isHttp ? (
          <>
            <Overview />
            <Upstreams />
            <Metrics />
          </>
        ) : (
          <>
            <Upstreams />
            <Metricstcp />
          </>
        )}
      </main>
      <Footer className="App-footer" />
    </div>
  );
}

export default App;
