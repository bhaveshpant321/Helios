// ==========================================================
// Helios History Page - Order History
// ==========================================================

const tableBody = document.getElementById("tableBody");
const filterType = document.getElementById("filterType");
const filterStatus = document.getElementById("filterStatus");

let allOrders = [];

// ==========================================================
// LOAD ORDER HISTORY
// ==========================================================
async function loadOrderHistory() {
  try {
    showLoading(true);
    
    const orders = await api.getUserOrders();
    
    allOrders = orders.map(order => ({
      id: order.id,
      symbol: order.symbol || 'N/A',
      side: order.side,
      type: order.type,
      quantity: parseFloat(order.quantity),
      filledQuantity: parseFloat(order.filled_quantity),
      price: order.price ? parseFloat(order.price) : null,
      status: order.status,
      createdAt: new Date(order.created_at)
    }));
    
    filterOrders();
    showLoading(false);
  } catch (error) {
    console.error('Error loading order history:', error);
    showError('Failed to load order history');
    showLoading(false);
  }
}

// ==========================================================
// RENDER ORDERS
// ==========================================================
function renderOrders(orders) {
  tableBody.innerHTML = "";
  
  if (orders.length === 0) {
    tableBody.innerHTML = '<tr><td colspan="9" style="text-align: center; padding: 2rem;">No orders found</td></tr>';
    return;
  }
  
  orders.forEach(order => {
    const row = document.createElement("tr");
    const sideIcon = order.side === 'BUY' 
      ? '<i class="fa-solid fa-arrow-up" style="color:#10b981;"></i>'
      : '<i class="fa-solid fa-arrow-down" style="color:#ef4444;"></i>';
    
    const priceDisplay = order.price ? '$' + order.price.toFixed(2) : 'Market';
    const total = order.price ? (order.quantity * order.price).toFixed(2) : 'N/A';
    const dateDisplay = order.createdAt.toLocaleString();
    
    const statusClass = order.status.toLowerCase().replace('_', '-');
    const statusDisplay = order.status.replace('_', ' ');
    
    let actionButton = '';
    if (order.status === 'OPEN' || order.status === 'PARTIALLY_FILLED') {
      actionButton = '<button class="btn" style="background:#ef4444; padding:0.3rem 0.8rem;" onclick="cancelOrder(' + order.id + ')">Cancel</button>';
    }
    
    row.innerHTML = '<td>' + order.id + '</td><td>' + order.symbol + '</td><td>' + sideIcon + ' ' + order.side + '</td><td>' + order.type + '</td><td>' + order.quantity.toFixed(8) + '</td><td>' + order.filledQuantity.toFixed(8) + '</td><td>' + priceDisplay + '</td><td>' + dateDisplay + '</td><td><span class="status ' + statusClass + '">' + statusDisplay + '</span>' + actionButton + '</td>';
    
    tableBody.appendChild(row);
  });
}

// ==========================================================
// FILTER ORDERS
// ==========================================================
function filterOrders() {
  let filtered = [...allOrders];
  
  if (filterType.value !== "all") {
    filtered = filtered.filter(o => o.side.toLowerCase() === filterType.value);
  }
  
  if (filterStatus.value !== "all") {
    const statusMap = {
      'open': ['OPEN', 'PARTIALLY_FILLED'],
      'completed': ['FILLED'],
      'cancelled': ['CANCELLED']
    };
    const statuses = statusMap[filterStatus.value] || [filterStatus.value.toUpperCase()];
    filtered = filtered.filter(o => statuses.includes(o.status));
  }
  
  renderOrders(filtered);
}

filterType.addEventListener("change", filterOrders);
filterStatus.addEventListener("change", filterOrders);

// ==========================================================
// CANCEL ORDER
// ==========================================================
async function cancelOrder(orderId) {
  if (!confirm('Are you sure you want to cancel this order?')) {
    return;
  }
  
  try {
    await api.cancelOrder(orderId);
    alert('Order cancelled successfully');
    await loadOrderHistory(); // Reload
  } catch (error) {
    console.error('Error cancelling order:', error);
    alert('Failed to cancel order: ' + (error.message || 'Unknown error'));
  }
}

// ==========================================================
// LOADING & ERROR STATES
// ==========================================================
function showLoading(show) {
  if (show) {
    tableBody.innerHTML = '<tr><td colspan="9" style="text-align: center; padding: 2rem;"><i class="fa-solid fa-spinner fa-spin" style="font-size: 2rem;"></i><p style="margin-top: 1rem;">Loading orders...</p></td></tr>';
  }
}

function showError(message) {
  tableBody.innerHTML = '<tr><td colspan="9" style="text-align: center; padding: 2rem; color: #ef4444;"><i class="fa-solid fa-triangle-exclamation" style="font-size: 2rem;"></i><p style="margin-top: 1rem;">' + message + '</p><button class="btn" onclick="loadOrderHistory()" style="margin-top: 1rem;">Retry</button></td></tr>';
}

// ==========================================================
// INITIALIZE
// ==========================================================
loadOrderHistory();
