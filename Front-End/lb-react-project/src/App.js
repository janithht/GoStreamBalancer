import './App.css';
import Header from './Components/Header/Header';
import Overview from './Components/Overview/Overview';
import Upstreams from './Components/Upstreams/Upstreams';
import RateLimit from './Components/RateLimit/Ratelimit';
import Footer from './Components/Footer/Footer';

function App() {
  return (
    <>
      <Header />
      <Overview />
      <Upstreams />
      <RateLimit />
      <Footer />
    </>
  );
}

export default App;
