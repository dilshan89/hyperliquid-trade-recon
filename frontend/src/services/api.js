import { API_BASE_URL } from '../config/config';

export const fetchPnLSummary = async () => {
  const response = await fetch(`${API_BASE_URL}/pnl`);
  if (!response.ok) {
    throw new Error('Failed to fetch P&L summary');
  }
  return response.json();
};

export const triggerRefresh = async (address, days) => {
  const url = `${API_BASE_URL}/refresh?address=${encodeURIComponent(address)}&days=${days}`;
  const response = await fetch(url, {
    method: 'POST',
  });
  if (!response.ok) {
    throw new Error('Failed to refresh data');
  }
  return response.json();
};