import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import InvoicesPage from "./components/InvoicesPage";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<InvoicesPage />} />
      </Routes>
    </Router>
  );
}

export default App;
