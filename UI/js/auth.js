// ==========================================================================
// LOGIN HANDLER
// ==========================================================================
const loginForm = document.getElementById("loginForm");
if (loginForm) {
  loginForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value.trim();
    const submitBtn = loginForm.querySelector('button[type="submit"]');
    
    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Logging in...';
    
    try {
      const response = await api.login(email, password);
      
      // Store JWT token and user info
      localStorage.setItem(CONFIG.STORAGE_KEYS.TOKEN, response.token);
      localStorage.setItem(CONFIG.STORAGE_KEYS.USER_ID, response.user.id);
      localStorage.setItem(CONFIG.STORAGE_KEYS.USERNAME, response.user.username);
      
      // Redirect to home
      window.location.href = "index.html";
    } catch (error) {
      console.error('Login error:', error);
      
      let errorMessage = "❌ Login failed. Please try again.";
      
      if (error instanceof APIError) {
        if (error.isAuthError()) {
          errorMessage = "❌ Invalid email or password";
        } else if (error.isNetworkError()) {
          errorMessage = "❌ Cannot connect to server. Please check if the API is running.";
        } else {
          errorMessage = `❌ ${error.message}`;
        }
      }
      
      alert(errorMessage);
      
      // Re-enable button
      submitBtn.disabled = false;
      submitBtn.innerHTML = '<i class="fa-solid fa-right-to-bracket"></i> Login';
    }
  });
}

// ==========================================================================
// REGISTER HANDLER
// ==========================================================================
const registerForm = document.getElementById("registerForm");
if (registerForm) {
  registerForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    
    const username = document.getElementById("name").value.trim();
    const email = document.getElementById("regEmail").value.trim();
    const password = document.getElementById("regPassword").value.trim();
    const submitBtn = registerForm.querySelector('button[type="submit"]');
    
    // Basic validation
    if (username.length < 3) {
      alert("⚠️ Username must be at least 3 characters");
      return;
    }
    
    if (password.length < 6) {
      alert("⚠️ Password must be at least 6 characters");
      return;
    }
    
    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Creating account...';
    
    try {
      const response = await api.register(username, email, password);
      
      // Store JWT token and user info
      localStorage.setItem(CONFIG.STORAGE_KEYS.TOKEN, response.token);
      localStorage.setItem(CONFIG.STORAGE_KEYS.USER_ID, response.user.id);
      localStorage.setItem(CONFIG.STORAGE_KEYS.USERNAME, response.user.username);
      
      // Show success and redirect
      alert("✅ Account created successfully!");
      window.location.href = "index.html";
    } catch (error) {
      console.error('Registration error:', error);
      
      let errorMessage = "❌ Registration failed. Please try again.";
      
      if (error instanceof APIError) {
        if (error.isValidationError()) {
          errorMessage = `⚠️ ${error.message}`;
        } else if (error.isNetworkError()) {
          errorMessage = "❌ Cannot connect to server. Please check if the API is running.";
        } else {
          errorMessage = `❌ ${error.message}`;
        }
      }
      
      alert(errorMessage);
      
      // Re-enable button
      submitBtn.disabled = false;
      submitBtn.innerHTML = '<i class="fa-solid fa-user-plus"></i> Register';
    }
  });
}

// ==========================================================================
// AUTH GUARD FOR PROTECTED PAGES
// ==========================================================================
const protectedPages = ["trade.html", "profile.html", "history.html"];
const currentPage = window.location.pathname.split("/").pop();

if (protectedPages.includes(currentPage)) {
  const token = localStorage.getItem(CONFIG.STORAGE_KEYS.TOKEN);
  if (!token) {
    window.location.href = "login.html";
  }
}

// ==========================================================================
// AUTO-REDIRECT IF LOGGED IN (FOR LOGIN/REGISTER PAGE)
// ==========================================================================
const authPages = ["login.html", "register.html"];
if (authPages.includes(currentPage)) {
  const token = localStorage.getItem(CONFIG.STORAGE_KEYS.TOKEN);
  if (token) {
    window.location.href = "index.html";
  }
}

// ==========================================================================
// LOGOUT FUNCTION (Call from profile page button)
// ==========================================================================
function logout() {
  localStorage.removeItem(CONFIG.STORAGE_KEYS.TOKEN);
  localStorage.removeItem(CONFIG.STORAGE_KEYS.USER_ID);
  localStorage.removeItem(CONFIG.STORAGE_KEYS.USERNAME);
  window.location.href = "login.html";
}

// ==========================================================================
// CHECK TOKEN VALIDITY
// ==========================================================================
function isAuthenticated() {
  return !!localStorage.getItem(CONFIG.STORAGE_KEYS.TOKEN);
}

function getUserId() {
  return parseInt(localStorage.getItem(CONFIG.STORAGE_KEYS.USER_ID));
}

function getUsername() {
  return localStorage.getItem(CONFIG.STORAGE_KEYS.USERNAME);
}
