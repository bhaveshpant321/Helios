// ==========================================================
// Helios Configuration
// ==========================================================

const CONFIG = {
  // API Base URL (Override via window.__HELIOS_CONFIG__.API_BASE_URL)
  API_BASE_URL: (window.__HELIOS_CONFIG__ && window.__HELIOS_CONFIG__.API_BASE_URL) || 'https://helios-api-ax7p.onrender.com/api/v1',
  
  // WebSocket URL (Override via window.__HELIOS_CONFIG__.WS_URL)
  WS_URL: (window.__HELIOS_CONFIG__ && window.__HELIOS_CONFIG__.WS_URL) || 'wss://helios-api-ax7p.onrender.com/ws/v1/market',
  
  // Local Storage Keys
  STORAGE_KEYS: {
    TOKEN: 'helios_token',
    USER_ID: 'helios_user_id',
    USERNAME: 'helios_username'
  },
  
  // API Endpoints
  ENDPOINTS: {
    // Auth
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    
    // Accounts
    BALANCES: '/account/balances',
    
    // Orders
    ORDERS: '/orders',
    ORDER_BY_ID: (id) => `/orders/${id}`,
    USER_ORDERS: '/orders/history',
    
    // Markets
    MARKETS: '/trading-pairs',
    ORDER_BOOK: (pairSymbol) => `/market/orderbook?pair=${pairSymbol}`,
    TRADES: (pairSymbol) => `/market/trades?pair=${pairSymbol}`,
  },
  
  // Request Timeout (ms)
  TIMEOUT: 10000,
  
  // Decimal places for display
  DECIMALS: {
    BTC: 8,
    ETH: 6,
    SOL: 4,
    USD: 2,
    USDT: 2,
    USDC: 2
  }
};

// Export for use in other files
if (typeof module !== 'undefined' && module.exports) {
  module.exports = CONFIG;
}
