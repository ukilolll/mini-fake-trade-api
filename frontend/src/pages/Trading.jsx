import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import axios from 'axios';
import '../styles/Trading.css';

export default function Trading() {
  const navigate = useNavigate();
  const { user, loading } = useContext(AuthContext);
  const [assets, setAssets] = useState([]);
  const [selectedAsset, setSelectedAsset] = useState(null);
  const [usdAmount, setUsdAmount] = useState('');
  const [loadingAssets, setLoadingAssets] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (!loading && !user) {
      // Allow access without login, but will redirect to login when trying to buy
    }
  }, [user, loading, navigate]);

  useEffect(() => {
    fetchAssets();
  }, []);

  const fetchAssets = async () => {
    setLoadingAssets(true);
    try {
      const response = await axios.get('http://localhost:3002/service', {
        withCredentials: true
      });
      if (response.data && Array.isArray(response.data)) {
        setAssets(response.data);
      }
    } catch (err) {
      setError('Failed to load assets');
      console.error('Error fetching assets:', err);
    } finally {
      setLoadingAssets(false);
    }
  };

  const handleBuy = async (e) => {
    e.preventDefault();
    if (!user) {
      navigate('/login');
      return;
    }

    if (!selectedAsset || !usdAmount || parseFloat(usdAmount) <= 0) {
      setError('Please select an asset and enter a valid USD amount');
      return;
    }

    try {
      const quantity = parseFloat(usdAmount) / selectedAsset.price;
      const response = await axios.post(
        'http://localhost:3002/trade/buy',
        {
          name: selectedAsset.asset_symbol,
          quantity: quantity
        },
        {
          headers: {
            'Content-Type': 'application/json'
          },
          withCredentials: true
        }
      );

      setSuccess(`Successfully bought ${quantity.toFixed(6)} shares of ${selectedAsset.asset_name} for $${usdAmount}`);
      setSelectedAsset(null);
      setUsdAmount('');
      setError('');
      
      // Refresh page after 1 second
      setTimeout(() => {
        window.location.reload();
      }, 100);
    } catch (err) {
      const status = err.response?.status;
      const message = err.response?.data?.msg || err.response?.data?.message;

      if (status === 400) {
        setError(message || 'bed request');
        return;
      }

      setError(message || 'Failed to buy stock');
      setSuccess('');
    }
  };

  if (loading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <div className="trading-container">
      <div className="trading-header">
        <h1>Trading Dashboard</h1>
        <div className="user-info">
          {user ? (
            <>
              <span>Welcome, {user?.username}</span>
              <span>Your Coin ${user?.coin.toFixed(6)}</span>
              <button 
                className="logout-btn"
                onClick={() => {
                  localStorage.removeItem('auth_token');
                  navigate('/login');
                }}
              >
                Logout
              </button>
              <button 
                className="portfolio-btn"
                onClick={() => navigate('/portfolio')}
              >
                My Portfolio
              </button>
            </>
          ) : (
            <button 
              className="login-btn"
              onClick={() => navigate('/login')}
            >
              Login
            </button>
          )}
        </div>
      </div>

      <div className="trading-content">
        <div className="buy-section">
          <h2>Buy Stock</h2>
          {error && <div className="error-message">{error}</div>}
          {success && <div className="success-message">{success}</div>}

          <form onSubmit={handleBuy} className="buy-form">
            <div className="form-group">
              <label htmlFor="asset">Select Stock:</label>
              <select
                id="asset"
                value={selectedAsset ? selectedAsset.asset_symbol : ''}
                onChange={(e) => {
                  const selected = assets.find(a => a.asset_symbol === e.target.value);
                  setSelectedAsset(selected);
                }}
                className="form-control"
              >
                <option value="">-- Choose a stock --</option>
                {assets.map((asset) => (
                  <option key={asset.asset_symbol} value={asset.asset_symbol}>
                    {asset.asset_symbol}
                  </option>
                ))}
              </select>
            </div>

            {selectedAsset && (
              <div className="asset-details">
                <p><strong>Current Price:</strong> ${selectedAsset.price.toFixed(2)}</p>
                <p><strong>Asset:</strong> {selectedAsset.asset_name}</p>
              </div>
            )}

            <div className="form-group">
              <label htmlFor="usdAmount">Amount (USD):</label>
              <input
                id="usdAmount"
                type="number"
                min="1"
                step="0.01"
                value={usdAmount}
                onChange={(e) => setUsdAmount(e.target.value)}
                className="form-control"
                placeholder="Enter USD amount"
              />
            </div>

            {selectedAsset && usdAmount && (
              <div className="total-cost">
                <strong>You will get:</strong> {(parseFloat(usdAmount) / selectedAsset.price).toFixed(6)} shares
              </div>
            )}

            <button type="submit" className="buy-btn" disabled={loadingAssets}>
              {loadingAssets ? 'Loading...' : 'Buy Stock'}
            </button>
          </form>
        </div>

        <div className="market-section">
          <h2>Available Stocks</h2>
          {loadingAssets && <div className="loading">Loading stocks...</div>}
          {assets.length === 0 && !loadingAssets && <p>No stocks available</p>}
          
          <div className="stock-grid">
            {assets.map((asset) => (
              <div key={asset.asset_symbol} className="stock-card">
                <h3>{asset.asset_symbol}</h3>
                <div className="stock-price">${asset.price.toFixed(2)}</div>
                <div className="stock-income">{asset.asset_name}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
