import React, { useState, useEffect, useCallback } from 'react';
import PnLTable from './components/PnLTable';
import TimeRangeSelector from './components/TimeRangeSelector';
import { fetchPnLSummary, triggerRefresh } from './services/api';
import { ACCOUNTS, DEFAULT_ACCOUNT, AUTO_REFRESH_INTERVAL_MS, DEFAULT_TIME_RANGE } from './config/config';
import './App.css';

function App() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshing, setRefreshing] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState(DEFAULT_ACCOUNT);
  const [selectedTimeRange, setSelectedTimeRange] = useState(DEFAULT_TIME_RANGE);

  // Fetch and load data for specific account and time range
  const fetchAndLoadData = useCallback(async (account, timeRange) => {
    try {
      setError(null);
      await triggerRefresh(account, timeRange);
      const result = await fetchPnLSummary();
      setData(result);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err.message);
    }
  }, []);

  // Handle manual refresh button click
  const handleRefresh = useCallback(async () => {
    try {
      setRefreshing(true);
      await fetchAndLoadData(selectedAccount, selectedTimeRange);
    } catch (err) {
      console.error('Error refreshing data:', err);
      setError(err.message);
    } finally {
      setRefreshing(false);
    }
  }, [selectedAccount, selectedTimeRange, fetchAndLoadData]);

  // Fetch data when account or time range changes
  useEffect(() => {
    let isCancelled = false;

    const fetchData = async () => {
      if (isCancelled) return;

      try {
        setRefreshing(true);
        await fetchAndLoadData(selectedAccount, selectedTimeRange);
      } catch (err) {
        if (!isCancelled) {
          console.error('Error fetching data:', err);
          setError(err.message);
        }
      } finally {
        if (!isCancelled) {
          setRefreshing(false);
        }
      }
    };

    // Initial fetch when filters change
    setLoading(true);
    fetchData().finally(() => setLoading(false));

    // Auto-refresh using configured interval
    const interval = setInterval(() => {
      if (!isCancelled) {
        fetchData();
      }
    }, AUTO_REFRESH_INTERVAL_MS);

    return () => {
      isCancelled = true;
      clearInterval(interval);
    };
  }, [selectedAccount, selectedTimeRange, fetchAndLoadData]);

  const handleAccountChange = (e) => {
    setSelectedAccount(e.target.value);
  };

  const handleTimeRangeChange = (range) => {
    setSelectedTimeRange(range);
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  return (
    <div className="App">
      <header className="App-header">
        <div className="top-bar">
          <div className="logo-section">
            <div className="logo">◆ Hyperliquid</div>
            <div className="separator">|</div>
            <div className="page-title">Trade Reconciliation</div>
          </div>
          <button
            onClick={handleRefresh}
            disabled={refreshing}
            className="refresh-button"
          >
            {refreshing ? 'Refreshing...' : 'Refresh Data'}
          </button>
        </div>
        <div className="header-controls">
          <div className="account-info">
            <span className="account-label">Select Account</span>
            <div className="select-wrapper">
              <select
                value={selectedAccount}
                onChange={handleAccountChange}
                className="account-select"
              >
                {ACCOUNTS.map((account) => (
                  <option key={account.address} value={account.address}>
                    {account.label} - {account.address}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <TimeRangeSelector
            selectedRange={selectedTimeRange}
            onRangeChange={handleTimeRangeChange}
          />
          <div className="pnl-info">
            <span className="pnl-label">Total P&L</span>
            <div className="pnl-value" style={{
              color: data?.totalPnL >= 0 ? '#22c55e' : '#ef4444',
            }}>
              {data ? formatCurrency(data.totalPnL) : '$0.00'}
            </div>
          </div>
        </div>
      </header>
      <main>
        <PnLTable data={data} loading={loading} error={error} />
      </main>
    </div>
  );
}

export default App;