// ==========================================================
// Helios Configuration
// ==========================================================

const CONFIG = {
  // API Base URL
  API_BASE_URL: 'http://localhost:8082/api/v1',
  
  // WebSocket URL
  // WS_URL: 'ws://localhost:8082/ws/v1/market',
  WS_URL: 'ws://localhost:8082/ws/v1/market',
  
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
