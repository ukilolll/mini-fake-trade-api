import React, { useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, AuthContext } from './context/AuthContext';
import Login from './pages/Login';
import Trading from './pages/Trading';
import Portfolio from './pages/Portfolio';
import './App.css';

function ProtectedRoute({ children }) {
  const { user, loading } = useContext(AuthContext);

  if (loading) {
    return <div className="loading-page">Loading...</div>;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

function AppContent() {
  const { user, loading } = useContext(AuthContext);

  // Handle Google OAuth callback
  React.useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');
    
    if (token) {
      // Cookie is already set by backend, just redirect
      window.history.replaceState({}, document.title, window.location.pathname);
      window.location.href = '/trading';
    }
  }, []);

  if (loading) {
    return <div className="loading-page">Loading...</div>;
  }

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route 
        path="/" 
        element={<Trading />}
      />
      <Route 
        path="/portfolio" 
        element={
          <ProtectedRoute>
            <Portfolio />
          </ProtectedRoute>
        } 
      />
    </Routes>
  );
}

function App() {
  return (
    <Router>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </Router>
  );
}

export default App;
