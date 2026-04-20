// ==========================================================
// Helios API Client
// Centralized API calls with error handling
// ==========================================================

class HeliosAPI {
  constructor() {
    this.baseURL = CONFIG.API_BASE_URL;
    this.timeout = CONFIG.TIMEOUT;
  }

  // ==========================================================
  // HELPER METHODS
  // ==========================================================

  /**
   * Get auth token from localStorage
   */
  getToken() {
    return localStorage.getItem(CONFIG.STORAGE_KEYS.TOKEN);
  }

  /**
   * Get auth headers with JWT token
   */
  getAuthHeaders() {
    const token = this.getToken();
    return {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    };
  }

  /**
   * Make API request with timeout and error handling
   */
  async request(endpoint, options = {}) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        ...options,
        headers: {
          ...this.getAuthHeaders(),
          ...options.headers
        },
        signal: controller.signal
      });

      clearTimeout(timeoutId);

      // Handle non-2xx responses
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new APIError(
          errorData.error || `HTTP ${response.status}: ${response.statusText}`,
          response.status,
          errorData
        );
      }

      return await response.json();
    } catch (error) {
      clearTimeout(timeoutId);
      
      if (error.name === 'AbortError') {
        throw new APIError('Request timeout - please try again', 408);
      }
      
      if (error instanceof APIError) {
        throw error;
      }
      
      throw new APIError(
        error.message || 'Network error - please check your connection',
        0,
        error
      );
    }
  }

  // ==========================================================
  // AUTHENTICATION
  // ==========================================================

  /**
   * Login user
   * @param {string} email 
   * @param {string} password 
   * @returns {Promise<{token: string, user: object}>}
   */
  async login(email, password) {
    return await this.request(CONFIG.ENDPOINTS.LOGIN, {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
  }

  /**
   * Register new user
   * @param {string} username 
   * @param {string} email 
   * @param {string} password 
   * @returns {Promise<{token: string, user: object}>}
   */
  async register(username, email, password) {
    return await this.request(CONFIG.ENDPOINTS.REGISTER, {
      method: 'POST',
      body: JSON.stringify({ username, email, password })
    });
  }

  // ==========================================================
  // ACCOUNTS
  // ==========================================================

  /**
   * Get user account balances
   * @returns {Promise<Array<{asset_id, ticker_symbol, name, balance, held_balance}>>}
   */
  async getBalances() {
    return await this.request(CONFIG.ENDPOINTS.BALANCES);
  }

  // ==========================================================
  // ORDERS
  // ==========================================================

  /**
   * Place a new order
   * @param {string} pairSymbol - Trading pair symbol (e.g., 'BTC/USD')
   * @param {string} side - 'BUY' or 'SELL'
   * @param {string} type - 'LIMIT' or 'MARKET'
   * @param {number|null} price - Required for LIMIT, null for MARKET
   * @param {number} quantity 
   * @returns {Promise<{order_id, message}>}
   */
  async placeOrder(pairSymbol, side, type, price, quantity) {
    return await this.request(CONFIG.ENDPOINTS.ORDERS, {
      method: 'POST',
      body: JSON.stringify({
        pair: pairSymbol,
        side,
        type,
        price: type === 'LIMIT' ? price : null,
        quantity
      })
    });
  }

  /**
   * Cancel an order
   * @param {number} orderId 
   * @returns {Promise<{message}>}
   */
  async cancelOrder(orderId) {
    return await this.request(CONFIG.ENDPOINTS.ORDER_BY_ID(orderId), {
      method: 'DELETE'
    });
  }

  /**
   * Get user's orders
   * @returns {Promise<Array<Order>>}
   */
  async getUserOrders() {
    return await this.request(CONFIG.ENDPOINTS.USER_ORDERS);
  }

  // ==========================================================
  // MARKETS
  // ==========================================================

  /**
   * Get all trading pairs (markets)
   * @returns {Promise<Array<{id, base_asset_id, quote_asset_id, symbol, base_name, quote_name}>>}
   */
  async getMarkets() {
    return await this.request(CONFIG.ENDPOINTS.MARKETS);
  }

  /**
   * Get order book for a trading pair
   * @param {string} pairSymbol - Trading pair symbol (e.g., 'BTC/USD')
   * @returns {Promise<{bids: Array, asks: Array}>}
   */
  async getOrderBook(pairSymbol) {
    return await this.request(CONFIG.ENDPOINTS.ORDER_BOOK(pairSymbol));
  }

  /**
   * Get trade history for a trading pair
   * @param {string} pairSymbol - Trading pair symbol (e.g., 'BTC/USD')
   * @param {number} limit - Optional, max number of trades
   * @returns {Promise<Array<Trade>>}
   */
  async getTrades(pairSymbol, limit = 50) {
    return await this.request(`${CONFIG.ENDPOINTS.TRADES(pairSymbol)}&limit=${limit}`);
  }
}

// ==========================================================
// CUSTOM ERROR CLASS
// ==========================================================

class APIError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }

  isAuthError() {
    return this.status === 401 || this.status === 403;
  }

  isValidationError() {
    return this.status === 400;
  }

  isNetworkError() {
    return this.status === 0;
  }
}

// ==========================================================
// SINGLETON INSTANCE
// ==========================================================

const api = new HeliosAPI();

// Export for use in other files
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { api, APIError };
}
