import React, { useContext, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import '../styles/Login.css';

export default function Login() {
  const navigate = useNavigate();
  const { user } = useContext(AuthContext);

  useEffect(() => {
    if (user) {
      navigate('/trading');
    }
  }, [user, navigate]);

  const handleGoogleLogin = () => {
    window.location.href = 'http://localhost:3002/auth/google/login';
  };

  return (
    <div className="login-container">
      <div className="login-box">
        <div className="login-header">
          <h1>Paper Trading</h1>
          <p>Practice trading with virtual money</p>
        </div>

        <div className="login-content">
          <button className="google-login-btn" onClick={handleGoogleLogin}>
            <svg className="google-icon" viewBox="0 0 24 24">
              <path fill="currentColor" d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"/>
            </svg>
            Sign in with Google
          </button>

          <div className="divider">
            <span>or</span>
          </div>

          <p className="info-text">
            No account needed. Just sign in with your Google account to get started.
          </p>
        </div>

        <div className="login-footer">
          <p>Your transactions are simulated. No real money involved.</p>
        </div>
      </div>
    </div>
  );
}
