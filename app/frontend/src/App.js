import { useEffect, useState } from 'react';
import logo from './logo.svg';
import './App.css';

function App() {
  const [ response, setResponse ] = useState("no response")

  useEffect( () => {
    setTimeout(() => fetch('/hello').then(res => res.text()).then(text => setResponse(text)), 10000)
  })

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
      <div>
        The response was {response}.
      </div>
    </div>
  );
}

export default App;
