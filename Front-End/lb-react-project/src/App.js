import './App.css';
import Header from './Components/Header/Header';
import Overview from './Components/Overview/Overview';
import Upstreams from './Components/Upstreams/Upstreams';
import Footer from './Components/Footer/Footer';
import Metrics from './Components/Metrics/Metrics';

function App() {
  return (
    <>
      <Header />
      <Overview />
      <Upstreams />
      <Metrics />
      <Footer />
    </>
  );
}

export default App;
