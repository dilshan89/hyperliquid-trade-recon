import React, { useState } from 'react';

const PnLTable = ({ data, loading, error }) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const formatCurrency = (value) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const getRowColor = (pnl) => {
    if (pnl > 0) return 'rgba(32,140,32,0.3)'; // Light green
    if (pnl < 0) return 'rgba(136,39,53,0.3)'; // Light pink
    return 'rgba(17, 24, 39, 0.3)';
  };

  if (loading) {
    return <div style={styles.message}>Loading trade data...</div>;
  }

  if (error) {
    return <div style={styles.error}>Error: {error}</div>;
  }

  if (!data || !data.dailyRecords || data.dailyRecords.length === 0) {
    return <div style={styles.message}>No trade data available</div>;
  }

  // Pagination calculations
  const totalRecords = data.dailyRecords.length;
  const totalPages = Math.ceil(totalRecords / rowsPerPage);
  const startIndex = (currentPage - 1) * rowsPerPage;
  const endIndex = Math.min(startIndex + rowsPerPage, totalRecords);
  const currentRecords = data.dailyRecords.slice(startIndex, endIndex);

  const handleRowsPerPageChange = (e) => {
    setRowsPerPage(Number(e.target.value));
    setCurrentPage(1); // Reset to first page
  };

  const handlePreviousPage = () => {
    if (currentPage > 1) {
      setCurrentPage(currentPage - 1);
    }
  };

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(currentPage + 1);
    }
  };

  return (
    <div style={styles.container}>
      <table style={styles.table}>
        <thead>
          <tr style={styles.headerRow}>
            <th style={{...styles.th, textAlign: 'left'}}>Date</th>
            <th style={{...styles.th, textAlign: 'center'}}>Number of Trades</th>
            <th style={{...styles.th, textAlign: 'right'}}>Daily P&L</th>
            <th style={{...styles.th, textAlign: 'right'}}>Cumulative P&L</th>
          </tr>
        </thead>
        <tbody>
          {currentRecords.map((record, index) => (
            <tr
              key={index}
              style={{
                ...styles.row,
                backgroundColor: getRowColor(record.dailyPnL),
              }}
            >
              <td style={{...styles.td, color: '#d1d5db', textAlign: 'left'}}>{record.date}</td>
              <td style={{...styles.td, textAlign: 'center'}}>{record.tradeCount}</td>
              <td style={{...styles.td, color: record.dailyPnL >= 0 ? '#4ade80' : '#fb7185', fontSize: '17px', textAlign: 'right'}}>
                {formatCurrency(record.dailyPnL)}
              </td>
              <td style={{...styles.td, color: record.cumulativePnL >= 0 ? '#4ade80' : '#fb7185', textAlign: 'right'}}>{formatCurrency(record.cumulativePnL)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination Controls */}
      <div style={styles.pagination}>
        <div style={styles.paginationLeft}>
          <span style={styles.paginationLabel}>Rows per page:</span>
          <select
            value={rowsPerPage}
            onChange={handleRowsPerPageChange}
            style={styles.rowsSelect}
          >
            <option value={5}>5</option>
            <option value={10}>10</option>
            <option value={25}>25</option>
            <option value={50}>50</option>
          </select>
        </div>

        <div style={styles.paginationRight}>
          <span style={styles.paginationInfo}>
            {startIndex + 1}-{endIndex} of {totalRecords}
          </span>
          <button
            onClick={handlePreviousPage}
            disabled={currentPage === 1}
            style={{
              ...styles.pageButton,
              opacity: currentPage === 1 ? 0.3 : 1,
              cursor: currentPage === 1 ? 'not-allowed' : 'pointer',
            }}
          >
            ‹
          </button>
          <button
            onClick={handleNextPage}
            disabled={currentPage === totalPages}
            style={{
              ...styles.pageButton,
              opacity: currentPage === totalPages ? 0.3 : 1,
              cursor: currentPage === totalPages ? 'not-allowed' : 'pointer',
            }}
          >
            ›
          </button>
        </div>
      </div>
    </div>
  );
};

const styles = {
  container: {
    padding: '20px',
  },
  table: {
    width: '100%',
    borderCollapse: 'separate',
    borderSpacing: '0',
    background: 'rgba(17, 24, 39, 0.75)',
    backdropFilter: 'blur(10px)',
    borderRadius: '16px',
    overflow: 'hidden',
    border: '1px solid rgba(34, 211, 238, 0.15)',
    boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
  },
  headerRow: {
    background: 'rgba(10, 14, 26, 0.6)',
    borderBottom: '1px solid rgba(34, 211, 238, 0.2)',
  },
  th: {
    padding: '16px 24px',
    fontWeight: '700',
    color: '#d1d5db',
    fontSize: '14px',
    textTransform: 'uppercase',
    letterSpacing: '1.2px',
  },
  row: {
    borderBottom: '1px solid rgba(55, 65, 81, 0.3)',
    transition: 'all 0.2s ease',
  },
  td: {
    padding: '16px 24px',
    color: '#ffffff',
    fontSize: '17px',
    fontFamily: "'Courier New', monospace",
  },
  message: {
    padding: '40px',
    textAlign: 'center',
    fontSize: '18px',
    color: '#9ca3af',
  },
  error: {
    padding: '40px',
    textAlign: 'center',
    fontSize: '18px',
    color: '#ef4444',
    background: 'rgba(239, 68, 68, 0.1)',
    borderRadius: '12px',
    border: '1px solid rgba(239, 68, 68, 0.2)',
  },
  pagination: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: '20px',
    padding: '10px 24px',
    background: 'rgba(17, 24, 39, 0.75)',
    backdropFilter: 'blur(10px)',
    borderRadius: '12px',
    border: '1px solid rgba(34, 211, 238, 0.15)',
  },
  paginationLeft: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  paginationLabel: {
    color: '#d1d5db',
    fontSize: '14px',
    fontWeight: '500',
  },
  rowsSelect: {
    background: 'rgba(10, 14, 26, 0.8)',
    color: '#e0f2fe',
    border: '1px solid rgba(34, 211, 238, 0.3)',
    borderRadius: '6px',
    padding: '6px 28px 6px 12px',
    fontSize: '14px',
    fontWeight: '500',
    cursor: 'pointer',
    outline: 'none',
    appearance: 'none',
    backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%2322d3ee' d='M6 9L1 4h10z'/%3E%3C/svg%3E")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'right 8px center',
  },
  paginationRight: {
    display: 'flex',
    alignItems: 'center',
    gap: '16px',
  },
  paginationInfo: {
    color: '#d1d5db',
    fontSize: '14px',
    fontWeight: '500',
  },
  pageButton: {
    background: 'rgba(34, 211, 238, 0.1)',
    color: '#22d3ee',
    border: '1px solid rgba(34, 211, 238, 0.3)',
    borderRadius: '6px',
    width: '36px',
    height: '36px',
    fontSize: '20px',
    fontWeight: '700',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'all 0.2s ease',
  },
};

export default PnLTable;