// ==========================================================
// Helios Home Page - Trading Pairs Overview
// ==========================================================

const stockGrid = document.getElementById("stockGrid");
const sortSelect = document.getElementById("sort");
const coinFilter = document.getElementById("coinFilter");

let allMarkets = [];
let filteredMarkets = [];

// ==========================================================
// LOAD MARKETS FROM API
// ==========================================================
async function loadMarkets() {
  try {
    showLoading(true);
    
    // Fetch all trading pairs
    allMarkets = await api.getMarkets();
    
    // Fetch order book for each pair to get current prices
    for (let market of allMarkets) {
      try {
        const orderBook = await api.getOrderBook(market.symbol);
        
        // Get best ask (lowest sell price) as current price
        if (orderBook.asks && orderBook.asks.length > 0) {
          market.currentPrice = parseFloat(orderBook.asks[0].price);
        } else if (orderBook.bids && orderBook.bids.length > 0) {
          market.currentPrice = parseFloat(orderBook.bids[0].price);
        } else {
          market.currentPrice = 0;
        }
        
        // Calculate 24h volume from recent trades
        const trades = await api.getTrades(market.symbol, 100);
        market.volume24h = trades.reduce((sum, trade) => {
          const tradeTime = new Date(trade.executed_at);
          const now = new Date();
          const hoursDiff = (now - tradeTime) / (1000 * 60 * 60);
          
          if (hoursDiff <= 24) {
            return sum + parseFloat(trade.quantity);
          }
          return sum;
        }, 0);
        
      } catch (error) {
        console.error(`Error loading data for ${market.symbol}:`, error);
        market.currentPrice = 0;
        market.volume24h = 0;
      }
    }
    
    // Populate filter dropdown
    populateFilters();
    
    // Initial render
    filterAndSort();
    showLoading(false);
    
  } catch (error) {
    console.error('Error loading markets:', error);
    showError('Failed to load markets. Please refresh the page.');
    showLoading(false);
  }
}

// ==========================================================
// POPULATE FILTER DROPDOWN
// ==========================================================
function populateFilters() {
  // Get unique base assets
  const baseAssets = [...new Set(allMarkets.map(m => m.base_name))];
  
  coinFilter.innerHTML = '<option value="all">All Trading Pairs</option>';
  baseAssets.forEach(asset => {
    const option = document.createElement('option');
    option.value = asset;
    option.textContent = asset;
    coinFilter.appendChild(option);
  });
}

// ==========================================================
// RENDER MARKETS TABLE
// ==========================================================
function renderMarkets(markets) {
  const tbody = stockGrid.querySelector("tbody");
  tbody.innerHTML = "";

  if (markets.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="6" style="text-align: center; padding: 2rem;">
          No trading pairs found
        </td>
      </tr>
    `;
    return;
  }

  markets.forEach((market) => {
    const row = document.createElement("tr");
    const priceDisplay = market.currentPrice > 0 
      ? `$${market.currentPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 8 })}`
      : 'N/A';
    
    const volumeDisplay = market.volume24h 
      ? market.volume24h.toFixed(4)
      : '0.0000';
    
    row.innerHTML = `
      <td><strong>${market.symbol}</strong></td>
      <td>${market.base_name}</td>
      <td>${market.quote_name}</td>
      <td>${priceDisplay}</td>
      <td>${volumeDisplay}</td>
      <td>
        <button class="btn" onclick="goToTrade(${market.id}, '${market.symbol}')">
          <i class="fa-solid fa-right-left"></i> Trade
        </button>
      </td>
    `;
    tbody.appendChild(row);
  });
}

// ==========================================================
// FILTER & SORT
// ==========================================================
function filterAndSort() {
  let filtered = [...allMarkets];

  // Filter by base asset
  if (coinFilter.value !== "all") {
    filtered = filtered.filter((m) => m.base_name === coinFilter.value);
  }

  // Sort by price
  if (sortSelect.value === "asc") {
    filtered.sort((a, b) => a.currentPrice - b.currentPrice);
  } else if (sortSelect.value === "desc") {
    filtered.sort((a, b) => b.currentPrice - a.currentPrice);
  }

  filteredMarkets = filtered;
  renderMarkets(filtered);
}

sortSelect.addEventListener("change", filterAndSort);
coinFilter.addEventListener("change", filterAndSort);

// ==========================================================
// TRADE REDIRECTION
// ==========================================================
function goToTrade(pairId, symbol) {
  localStorage.setItem("selectedTradingPair", JSON.stringify({ id: pairId, symbol: symbol }));
  window.location.href = "trade.html";
}

// ==========================================================
// LOADING & ERROR STATES
// ==========================================================
function showLoading(show) {
  const tbody = stockGrid.querySelector("tbody");
  if (show) {
    tbody.innerHTML = `
      <tr>
        <td colspan="6" style="text-align: center; padding: 2rem;">
          <i class="fa-solid fa-spinner fa-spin" style="font-size: 2rem;"></i>
          <p style="margin-top: 1rem;">Loading markets...</p>
        </td>
      </tr>
    `;
  }
}

function showError(message) {
  const tbody = stockGrid.querySelector("tbody");
  tbody.innerHTML = `
    <tr>
      <td colspan="6" style="text-align: center; padding: 2rem; color: #ef4444;">
        <i class="fa-solid fa-triangle-exclamation" style="font-size: 2rem;"></i>
        <p style="margin-top: 1rem;">${message}</p>
        <button class="btn" onclick="loadMarkets()" style="margin-top: 1rem;">
          <i class="fa-solid fa-rotate"></i> Retry
        </button>
      </td>
    </tr>
  `;
}

// ==========================================================
// INITIALIZE
// ==========================================================
loadMarkets();
