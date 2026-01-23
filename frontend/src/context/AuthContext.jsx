import React, { createContext, useState, useEffect } from 'react';
import axios from 'axios';

// Configure axios to send cookies with all requests
axios.defaults.withCredentials = true;

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

      // Check if user is logged in using /auth/check endpoint
    const checkAuth = async () => {
      try {
        // Call /auth/check endpoint - cookie will be sent automatically
        const response = await axios.get(
          'http://localhost:3002/auth/user/profile',
          {},
          {
            headers: {
              'Content-Type': 'application/json'
            }
          }
        );

        // Status 200 = logged in
        if (response.status === 200 && response.data) {
          setUser({
            username: response.data.username,
            coin: response.data.coin
          });
        }
      } catch (error) {
        // Status 401 = not logged in or token invalid
        if (error.response && error.response.status === 401) {
          console.log('User not authenticated');
          setUser(null);
        } else {
          console.error('Auth check error:', error);
          setUser(null);
        }
      } finally {
        setLoading(false);
      }
    };

  useEffect(() => {


    checkAuth();
  }, []);

  const logout = async() => {
    try {
    await axios.post(
      'http://localhost:3002/auth/logout',
      {},
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
    setUser(null);
    }catch(err){
      console.error(err)
    }
  };

  const loadUser = async() => {
    checkAuth()
  }

  return (
    <AuthContext.Provider value={{ user, setUser, loading, logout ,loadUser }}>
      {children}
    </AuthContext.Provider>
  );
};
