import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import axios from 'axios';
import '../styles/Portfolio.css';

export default function Portfolio() {
  const navigate = useNavigate();
  const { user, loading } = useContext(AuthContext);
  const [portfolio, setPortfolio] = useState([]);
  const [loadingPortfolio, setLoadingPortfolio] = useState(false);
  const [error, setError] = useState('');
  const [sellAsset, setSellAsset] = useState(null);
  const [sellQuantity, setSellQuantity] = useState('');
  const wsRef = React.useRef(null);

  useEffect(() => {
    if (!loading && !user) {
      navigate('/login');
    }
  }, [user, loading, navigate]);

  useEffect(() => {
    fetchPortfolio();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const fetchPortfolio = async () => {
    setLoadingPortfolio(true);
    try {
      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
      const wsUrl = `${protocol}://localhost:3002/trade/check/assets`;
      
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connected');
      };

      wsRef.current.onmessage = (event) => {
        try {
          const parsedData = JSON.parse(event.data);
          console.log('WebSocket message:', parsedData);
          
          if (parsedData.type === 'error') {
            setError(parsedData.error || 'Failed to load portfolio');
            setPortfolio([]);
            setLoadingPortfolio(false);
            return;
          }

          // Handle initial buy price and quantity data (one-time)
          if (parsedData.type === 'not-real-time') {
            const initialData = Array.isArray(parsedData.data) ? parsedData.data : [];
            console.log('Initial portfolio data:', initialData);
            
            setPortfolio(prevPortfolio => {
              // If portfolio exists, merge initial data with existing real-time data
              if (prevPortfolio.length > 0) {
                return prevPortfolio.map((asset) => {
                  const initialAsset = initialData.find(item => item.s === asset.s);
                  if (initialAsset) {
                    return {
                      ...asset,
                      buy_price: initialAsset.buy_price,
                      quantity: initialAsset.quantity
                    };
                  }
                  return asset;
                });
              }
              
              // If no portfolio yet, store initial data for later merge
              return initialData.map(item => ({
                s: item.s,
                stock_price_now: 0,
                income: '',
                buy_price: item.buy_price,
                quantity: item.quantity
              }));
            });
            setLoadingPortfolio(false);
            setError('');
          } 
          // Handle real-time stock price updates (continuous)
          else if (parsedData.type === 'real-time') {
            const realTimeData = Array.isArray(parsedData.data) ? parsedData.data : [];
            
            setPortfolio(prevPortfolio => {
              if (prevPortfolio.length === 0) {
                // If portfolio is empty, initialize from real-time data
                return realTimeData.map(item => ({
                  s: item.s,
                  stock_price_now: item.stock_price_now,
                  income: item.income,
                  buy_price: 0,
                  quantity: 0
                }));
              } else {
                // Merge real-time data with existing portfolio by matching stock symbol
                return prevPortfolio.map((asset) => {
                  const rtData = realTimeData.find(rt => rt.s === asset.s);
                  if (rtData) {
                    return {
                      ...asset,
                      stock_price_now: rtData.stock_price_now,
                      income: rtData.income
                    };
                  }
                  return asset;
                });
              }
            });
            setLoadingPortfolio(false);
            setError('');
          }
        } catch (err) {
          console.error('Error parsing portfolio data:', err);
          setError('Failed to parse portfolio data');
          setLoadingPortfolio(false);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Failed to connect to portfolio data');
        setLoadingPortfolio(false);
      };

      wsRef.current.onclose = () => {
        console.log('WebSocket disconnected');
      };
    } catch (err) {
      setError('Failed to load portfolio');
      console.error('Error connecting to portfolio:', err);
      setLoadingPortfolio(false);
    }
  };

  const handleSell = async (asset) => {
    if (!sellQuantity || parseFloat(sellQuantity) <= 0) {
      setError('Please enter a valid quantity');
      return;
    }

    if (parseFloat(sellQuantity) > (asset.quantity || 0)) {
      setError('Quantity exceeds your holdings');
      return;
    }

    try {
      const response = await axios.post(
        'http://localhost:3002/trade/sell',
        {
          name: asset.s,
          quantity: parseFloat(sellQuantity)
        },
        {
          headers: {
            'Content-Type': 'application/json'
          },
          withCredentials: true
        }
      );

      setError('');
      setSellAsset(null);
      setSellQuantity('');
      
      setTimeout(() => {
        window.location.reload();
      }, 100);
    } catch (err) {
        const status = err.response?.status;
        if (status === 400) {
        setError(message || 'bed request');
        return;
      }

      setError(err.response?.data?.message || 'Failed to sell stock');
    }
  };

  const calculateTotalValue = () => {
    return portfolio.reduce((total, asset) => {
      return total + ((asset.quantity || 0) * (asset.stock_price_now || 0));
    }, 0);
  };

  const calculateTotalGain = () => {
    return portfolio.reduce((total, asset) => {
      const value = ((asset.quantity || 0) * (asset.stock_price_now || 0));
      const cost = ((asset.quantity || 0) * (asset.buy_price || 0));
      return total + (value - cost);
    }, 0);
  };

  const handleResetTreading  = async() => {
    try{
      await axios.delete(
        'http://localhost:3002/trade/reset',
        {
          withCredentials: true
        }
      );
    }catch(err){
      console.error(err)
    }
  }

  if (loading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <div className="portfolio-container">
      <div className="portfolio-header">
        <h1>My Portfolio</h1>
        <div className="user-actions">
          <button 
            className="trading-btn"
            onClick={() => navigate('/')}
          >
            Back to Trading
          </button>
          <button 
            className="logout-btn"
            onClick={async () => {
              await handleResetTreading()
              navigate('/')
            }}
          >
            reset
          </button>
        </div>
      </div>

      <div className="portfolio-stats">
        <div className="stat-card">
          <h3>Total Portfolio Value</h3>
          <p className="stat-value">${calculateTotalValue().toFixed(2)}</p>
        </div>
        <div className="stat-card">
          <h3>Total Gain/Loss</h3>
          <p className={`stat-value ${calculateTotalGain() >= 0 ? 'positive' : 'negative'}`}>
            {calculateTotalGain() >= 0 ? '+' : ''} ${calculateTotalGain().toFixed(2)}
          </p>
        </div>
      </div>

      <div className="portfolio-content">
        {error && <div className="error-message">{error}</div>}

        {loadingPortfolio && <div className="loading">Loading portfolio...</div>}

        {portfolio.length === 0 && !loadingPortfolio && (
          <div className="empty-state">
            <p>You don't own any stocks yet.</p>
            <button className="trading-btn" onClick={() => navigate('/')}>
              Start Trading
            </button>
          </div>
        )}

        {portfolio.length > 0 && (
          <div className="portfolio-table-container">
            <table className="portfolio-table">
              <thead>
                <tr>
                  <th>Stock</th>
                  <th>Quantity</th>
                  <th>Buy Price</th>
                  <th>Current Price</th>
                  <th>Total Value</th>
                  <th>Gain/Loss</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {portfolio.map((asset) => {
                  const totalValue = (asset.quantity || 0) * (asset.stock_price_now || 0);
                  const totalCost = (asset.quantity || 0) * (asset.buy_price || 0);
                  const gainLoss = totalValue - totalCost;
                  const gainLossPercent = totalCost > 0 ? ((gainLoss / totalCost) * 100) : 0;

                  return (
                    <tr key={asset.s}>
                      <td className="stock-name">{asset.s}</td>
                      <td>{(asset.quantity || 0).toFixed(6)}</td>
                      <td>${(asset.buy_price || 0).toFixed(2)}</td>
                      <td>${(asset.stock_price_now || 0).toFixed(2)}</td>
                      <td>${totalValue.toFixed(2)}</td>
                      <td className={gainLoss >= 0 ? 'positive' : 'negative'}>
                        {gainLoss >= 0 ? '+' : ''} ${gainLoss.toFixed(2)} ({gainLossPercent.toFixed(2)}%)
                      </td>
                      <td>
                        <button 
                          className="sell-btn-small"
                          onClick={() => setSellAsset(asset)}
                        >
                          Sell
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {sellAsset && (
        <div className="modal-overlay" onClick={() => setSellAsset(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>Sell {sellAsset.s}</h2>
            <div className="modal-info">
              <p><strong>Available:</strong> {(sellAsset.quantity || 0).toFixed(2)} shares</p>
              <p><strong>Current Price:</strong> ${(sellAsset.stock_price_now || 0).toFixed(2)}</p>
            </div>

            <div className="form-group">
              <div className="quantity-input-wrapper">
                <div>
                  <label htmlFor="sell-quantity">Quantity to Sell:</label>
                  <input
                    id="sell-quantity"
                    type="number"
                    min="1"
                    step="0.01"
                    max={sellAsset.quantity || 0}
                    value={sellQuantity}
                    onChange={(e) => setSellQuantity(e.target.value)}
                    className="form-control"
                    placeholder="Enter quantity"
                  />
                </div>
                <button 
                  className="sell-all-btn"
                  onClick={() => setSellQuantity(sellAsset.quantity.toString())}
                >
                  Sell All
                </button>
              </div>
            </div>

            {sellQuantity && (
              <div className="total-cost">
                <strong>Total Proceeds:</strong> ${(parseFloat(sellQuantity) * (sellAsset.stock_price_now || 0)).toFixed(6)}
              </div>
            )}

            <div className="modal-actions">
              <button 
                className="sell-btn"
                onClick={() => handleSell(sellAsset)}
              >
                Sell
              </button>
              <button 
                className="cancel-btn"
                onClick={() => setSellAsset(null)}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
