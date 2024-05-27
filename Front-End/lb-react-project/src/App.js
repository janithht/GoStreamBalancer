import './App.css';
import Header from './Components/Header/Header';
import Overview from './Components/Overview/Overview';
import Upstreams from './Components/Upstreams/Upstreams';
import Footer from './Components/Footer/Footer';
import Metrics from './Components/Metrics/Metrics';

function App() {
  return (
    <div className="App">
      <Header className="App-header" />
      <main className="App-main">
        <Overview />
        <Upstreams />
        <Metrics />
      </main>
      <Footer className="App-footer" />
    </div>
  );
}

export default App;
