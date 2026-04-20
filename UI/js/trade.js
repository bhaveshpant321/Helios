// ==========================================================
// Helios Trade Page - Order Book & Trading
// ==========================================================

let currentTradingPair = null;
let ws = null;
let orderBook = { bids: [], asks: [] };

// ==========================================================
// INITIALIZE PAGE
// ==========================================================
document.addEventListener('DOMContentLoaded', async () => {
  // Get selected trading pair
  const storedPair = localStorage.getItem("selectedTradingPair");
  if (!storedPair) {
    alert("No trading pair selected");
    window.location.href = "index.html";
    return;
  }

  currentTradingPair = JSON.parse(storedPair);
  
  // Set page title
  document.getElementById('pairSymbol').textContent = currentTradingPair.symbol;
  
  // Load initial data
  await loadOrderBook();
  await loadUserBalances();
  await loadRecentTrades();
  
  // Connect WebSocket for real-time updates
  connectWebSocket();
  
  // Setup event listeners
  setupOrderForm();
});

// ==========================================================
// LOAD ORDER BOOK
// ==========================================================
async function loadOrderBook() {
  try {
    const data = await api.getOrderBook(currentTradingPair.symbol);
    orderBook = data;
    renderOrderBook();
    updateBestPrices();
  } catch (error) {
    console.error('Error loading order book:', error);
    showNotification('Failed to load order book', 'error');
  }
}

// ==========================================================
// RENDER ORDER BOOK
// ==========================================================
function renderOrderBook() {
  const asksContainer = document.getElementById('asks');
  const bidsContainer = document.getElementById('bids');
  
  // Render asks (sell orders) - lowest first
  asksContainer.innerHTML = '';
  const sortedAsks = [...orderBook.asks].sort((a, b) => parseFloat(a.price) - parseFloat(b.price));
  sortedAsks.slice(0, 10).forEach(ask => {
    const row = document.createElement('div');
    row.className = 'orderbook-row ask';
    row.innerHTML = `
      <span class="price">${parseFloat(ask.price).toFixed(2)}</span>
      <span class="quantity">${parseFloat(ask.total_quantity).toFixed(8)}</span>
    `;
    row.onclick = () => fillOrderForm(parseFloat(ask.price));
    asksContainer.appendChild(row);
  });
  
  // Render bids (buy orders) - highest first
  bidsContainer.innerHTML = '';
  const sortedBids = [...orderBook.bids].sort((a, b) => parseFloat(b.price) - parseFloat(a.price));
  sortedBids.slice(0, 10).forEach(bid => {
    const row = document.createElement('div');
    row.className = 'orderbook-row bid';
    row.innerHTML = `
      <span class="price">${parseFloat(bid.price).toFixed(2)}</span>
      <span class="quantity">${parseFloat(bid.total_quantity).toFixed(8)}</span>
    `;
    row.onclick = () => fillOrderForm(parseFloat(bid.price));
    bidsContainer.appendChild(row);
  });
}

// ==========================================================
// UPDATE BEST PRICES
// ==========================================================
function updateBestPrices() {
  const bestAsk = orderBook.asks.length > 0 
    ? Math.min(...orderBook.asks.map(a => parseFloat(a.price)))
    : 0;
  
  const bestBid = orderBook.bids.length > 0
    ? Math.max(...orderBook.bids.map(b => parseFloat(b.price)))
    : 0;
  
  const spread = bestAsk && bestBid ? (bestAsk - bestBid).toFixed(2) : '0.00';
  
  document.getElementById('bestAsk').textContent = bestAsk ? `$${bestAsk.toFixed(2)}` : 'N/A';
  document.getElementById('bestBid').textContent = bestBid ? `$${bestBid.toFixed(2)}` : 'N/A';
  document.getElementById('spread').textContent = `$${spread}`;
}

// ==========================================================
// LOAD USER BALANCES
// ==========================================================
async function loadUserBalances() {
  try {
    const balances = await api.getBalances();
    
    // Display balances
    const balancesHtml = balances.map(bal => `
      <div class="balance-item">
        <span><strong>${bal.ticker_symbol}:</strong></span>
        <span>${parseFloat(bal.balance).toFixed(8)}</span>
      </div>
    `).join('');
    
    document.getElementById('userBalances').innerHTML = balancesHtml;
  } catch (error) {
    console.error('Error loading balances:', error);
    showNotification('Failed to load balances', 'error');
  }
}

// ==========================================================
// LOAD RECENT TRADES
// ==========================================================
async function loadRecentTrades() {
  try {
     if (!currentTradingPair || !currentTradingPair.symbol) {

      console.error('No trading pair selected for trades');

      return;
    }
    console.log('Loading trades for:', currentTradingPair.symbol);
    const trades = await api.getTrades(currentTradingPair.symbol, 20);
    
    const tradesContainer = document.getElementById('recentTrades');
    tradesContainer.innerHTML = '';
    
    trades.forEach(trade => {
      const row = document.createElement('div');
      row.className = 'trade-row';
      const time = new Date(trade.executed_at).toLocaleTimeString();
      row.innerHTML = `
        <span>${time}</span>
        <span>$${parseFloat(trade.price).toFixed(2)}</span>
        <span>${parseFloat(trade.quantity).toFixed(8)}</span>
      `;
      tradesContainer.appendChild(row);
    });
  } catch (error) {
    console.error('Error loading trades:', error);
  }
}

// ==========================================================
// WEBSOCKET CONNECTION
// ==========================================================
function connectWebSocket() {
  try {
    // URL encode the trading pair symbol (BTC/USD -> BTC%2FUSD)
    const encodedSymbol = encodeURIComponent(currentTradingPair.symbol);
    const wsUrl = `${CONFIG.WS_URL}/${encodedSymbol}`;
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      showNotification('Connected to live updates', 'success');
    };
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        
        // Update order book based on message type
        if (message.type === 'orderbook' && message.data) {
          orderBook = message.data;
          renderOrderBook();
          updateBestPrices();
        }
      } catch (error) {
        console.error('WebSocket message error:', error);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      showNotification('Live updates disconnected', 'error');
    };
    
    ws.onclose = () => {
      console.log('WebSocket closed');
      // Attempt reconnect after 5 seconds
      setTimeout(connectWebSocket, 5000);
    };
  } catch (error) {
    console.error('WebSocket connection error:', error);
  }
}

// ==========================================================
// ORDER FORM SETUP
// ==========================================================
function setupOrderForm() {
  const orderTypeRadios = document.getElementsByName('orderType');
  const priceInput = document.getElementById('price');
  const quantityInput = document.getElementById('quantity');
  const totalDisplay = document.getElementById('total');
  
  // Toggle price input based on order type
  orderTypeRadios.forEach(radio => {
    radio.addEventListener('change', (e) => {
      if (e.target.value === 'MARKET') {
        priceInput.disabled = true;
        priceInput.value = '';
        priceInput.placeholder = 'Market Price';
      } else {
        priceInput.disabled = false;
        priceInput.placeholder = 'Enter price';
      }
      calculateTotal();
    });
  });
  
  // Calculate total on input change
  priceInput.addEventListener('input', calculateTotal);
  quantityInput.addEventListener('input', calculateTotal);
  
  // Buy/Sell button handlers
  document.getElementById('buyBtn').addEventListener('click', () => placeOrder('BUY'));
  document.getElementById('sellBtn').addEventListener('click', () => placeOrder('SELL'));
}

// ==========================================================
// CALCULATE TOTAL
// ==========================================================
function calculateTotal() {
  const orderType = document.querySelector('input[name="orderType"]:checked').value;
  const quantity = parseFloat(document.getElementById('quantity').value) || 0;
  const totalDisplay = document.getElementById('total');
  
  if (orderType === 'LIMIT') {
    const price = parseFloat(document.getElementById('price').value) || 0;
    const total = price * quantity;
    totalDisplay.textContent = `$${total.toFixed(2)}`;
  } else {
    totalDisplay.textContent = 'Market Price';
  }
}

// ==========================================================
// FILL ORDER FORM FROM ORDER BOOK CLICK
// ==========================================================
function fillOrderForm(price) {
  document.getElementById('limitRadio').checked = true;
  document.getElementById('price').disabled = false;
  document.getElementById('price').value = price;
  calculateTotal();
}

// ==========================================================
// PLACE ORDER
// ==========================================================
async function placeOrder(side) {
  const orderType = document.querySelector('input[name="orderType"]:checked').value;
  const quantity = parseFloat(document.getElementById('quantity').value);
  const price = orderType === 'LIMIT' ? parseFloat(document.getElementById('price').value) : null;
  
  // Validation
  if (!quantity || quantity <= 0) {
    showNotification('Please enter a valid quantity', 'error');
    return;
  }
  
  if (orderType === 'LIMIT' && (!price || price <= 0)) {
    showNotification('Please enter a valid price', 'error');
    return;
  }
  
  const btn = side === 'BUY' ? document.getElementById('buyBtn') : document.getElementById('sellBtn');
  const originalText = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Placing...';
  
  try {
    const response = await api.placeOrder(
      currentTradingPair.symbol,
      side,
      orderType,
      price,
      quantity
    );
    
    showNotification(`Order placed successfully! ID: ${response.order_id}`, 'success');
    
    // Clear form
    document.getElementById('price').value = '';
    document.getElementById('quantity').value = '';
    document.getElementById('total').textContent = '$0.00';
    
    // Refresh data
    await loadOrderBook();
    await loadUserBalances();
    await loadRecentTrades();
    
  } catch (error) {
    console.error('Error placing order:', error);
    let errorMsg = 'Failed to place order';
    
    if (error instanceof APIError) {
      errorMsg = error.message;
    }
    
    showNotification(errorMsg, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = originalText;
  }
}

// ==========================================================
// NOTIFICATION SYSTEM
// ==========================================================
function showNotification(message, type = 'info') {
  const notification = document.createElement('div');
  notification.className = `notification ${type}`;
  notification.innerHTML = `
    <i class="fa-solid ${type === 'success' ? 'fa-check-circle' : type === 'error' ? 'fa-exclamation-circle' : 'fa-info-circle'}"></i>
    <span>${message}</span>
  `;
  
  document.body.appendChild(notification);
  
  setTimeout(() => {
    notification.classList.add('show');
  }, 100);
  
  setTimeout(() => {
    notification.classList.remove('show');
    setTimeout(() => notification.remove(), 300);
  }, 3000);
}

// ==========================================================
// CLEANUP ON PAGE UNLOAD
// ==========================================================
window.addEventListener('beforeunload', () => {
  if (ws) {
    ws.close();
  }
});
