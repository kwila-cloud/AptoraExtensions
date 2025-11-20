import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import "./App.css";

function App() {
  return (
    <Router>
      <div className="App">
        <h1>Aptora Extensions</h1>
        <Routes>
          <Route path="/" element={<div>Welcome to Aptora Extensions</div>} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
