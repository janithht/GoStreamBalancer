import './App.css';
import Main from './Components/Main';
import Header from './Components/Header/Header';
import Overview from './Components/Overview/Overview';
import Upstreams from './Components/Upstreams/Upstreams';
import RateLimit from './Components/RateLimit/Ratelimit';
import Footer from './Components/Footer/Footer';

function App() {
  return (
    <div>
      <Main/>
      <Header />
      <Overview />
      <Upstreams />
      <RateLimit />
      <Footer />
    </div>
  );
}

export default App;
