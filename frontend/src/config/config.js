// API Configuration
// Use relative URL for production (embedded), absolute URL for development
export const API_BASE_URL = process.env.NODE_ENV === 'production'
  ? '/api'  // Production: relative URL (same server)
  : 'http://localhost:8080/api';  // Development: absolute URL (CORS)

// Polling Configuration
// With incremental caching, we can refresh more frequently without rate limiting
// Only new trades are fetched after the initial load
export const AUTO_REFRESH_INTERVAL_MS = 10000; // 10 seconds (safe with incremental caching)

// Account Configuration
export const ACCOUNTS = [
  { address: '0x20c2d95a3dfdca9e9ad12794d5fa6fad99da44f5', label: 'Account 1' },
  { address: '0xb83de012dba672c76a7dbbbf3e459cb59d7d6e36', label: 'Account 2' },
  { address: '0x091144e651b334341eabdbbbfed644ad0100023e', label: 'Account 3' },
];

export const DEFAULT_ACCOUNT = ACCOUNTS[0].address;

// Time Range Configuration
export const TIME_RANGES = [
  { value: 1, label: '24H' },
  { value: 7, label: '7D' },
  { value: 30, label: '30D' },
  { value: 90, label: '90D' },
];

export const DEFAULT_TIME_RANGE = 30; // 30 days