# Production-Ready Frontend Authentication with HTTP-Only Cookies

## Overview
Your backend uses HTTP-only cookies for JWT storage, which is secure against XSS attacks. The frontend doesn't directly access the token - it's automatically sent with requests.

---

## Key Concepts

### ✅ What HTTP-Only Cookies Do
- **Automatically sent** with every request to your API domain
- **Secure** - JavaScript cannot access them (prevents XSS)
- **Backend managed** - Set via `Set-Cookie` header

### ⚠️ Frontend Responsibilities
- Maintain auth state (user data, isAuthenticated)
- Handle auth redirects
- Validate session on app load
- Clear state on logout

---

## Implementation by Framework

### 1. React + Context API (Recommended for Medium Apps)

#### `src/context/AuthContext.jsx`
```jsx
import { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

// Configure axios to send cookies
axios.defaults.withCredentials = true;
axios.defaults.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // Validate session on mount
  useEffect(() => {
    validateSession();
  }, []);

  const validateSession = async () => {
    try {
      const response = await axios.get('/auth/validate');
      if (response.data.message === 'success') {
        setUser(response.data.data);
        setIsAuthenticated(true);
      }
    } catch (error) {
      console.error('Session validation failed:', error);
      setUser(null);
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email, password) => {
    try {
      const response = await axios.post('/auth/login', { email, password });
      
      if (response.data.message === 'success') {
        setUser(response.data.data);
        setIsAuthenticated(true);
        return { success: true, data: response.data.data };
      }
      
      return { success: false, message: 'Login failed' };
    } catch (error) {
      const message = error.response?.data?.message || 'Login failed';
      return { success: false, message };
    }
  };

  const logout = async () => {
    try {
      await axios.get('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear state regardless of API response
      setUser(null);
      setIsAuthenticated(false);
    }
  };

  const register = async (userData) => {
    try {
      const response = await axios.post('/auth/register', userData);
      return { success: true, data: response.data.data };
    } catch (error) {
      const message = error.response?.data?.message || 'Registration failed';
      return { success: false, message };
    }
  };

  const verifyOTP = async (id, email, otp) => {
    try {
      const response = await axios.post('/auth/verify-otp', { id, email, otp });
      return { success: true, message: response.data.data };
    } catch (error) {
      const message = error.response?.data?.message || 'OTP verification failed';
      return { success: false, message };
    }
  };

  const value = {
    user,
    isAuthenticated,
    isLoading,
    login,
    logout,
    register,
    verifyOTP,
    validateSession,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};
```

#### `src/components/ProtectedRoute.jsx`
```jsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const ProtectedRoute = ({ children, requiredRole }) => {
  const { isAuthenticated, isLoading, user } = useAuth();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole && user?.role !== requiredRole) {
    return <Navigate to="/unauthorized" replace />;
  }

  return children;
};
```

#### `src/App.jsx`
```jsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import VerifyOTP from './pages/VerifyOTP';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/verify-otp/:id/:email" element={<VerifyOTP />} />
          
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="/admin"
            element={
              <ProtectedRoute requiredRole="admin">
                <AdminPanel />
              </ProtectedRoute>
            }
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
```

#### Usage Example - `src/pages/Login.jsx`
```jsx
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    const result = await login(email, password);
    
    if (result.success) {
      navigate('/dashboard');
    } else {
      setError(result.message);
    }
    
    setIsLoading(false);
  };

  return (
    <div className="max-w-md mx-auto mt-8">
      <form onSubmit={handleSubmit} className="bg-white shadow-md rounded px-8 pt-6 pb-8">
        <h2 className="text-2xl mb-6 font-bold">Login</h2>
        
        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}
        
        <div className="mb-4">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="shadow appearance-none border rounded w-full py-2 px-3"
            required
          />
        </div>
        
        <div className="mb-6">
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="shadow appearance-none border rounded w-full py-2 px-3"
            required
          />
        </div>
        
        <button
          type="submit"
          disabled={isLoading}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded w-full"
        >
          {isLoading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  );
};

export default Login;
```

---

### 2. React + Redux Toolkit (For Large Apps)

#### `src/features/auth/authSlice.js`
```js
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';

axios.defaults.withCredentials = true;
axios.defaults.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export const validateSession = createAsyncThunk(
  'auth/validateSession',
  async (_, { rejectWithValue }) => {
    try {
      const response = await axios.get('/auth/validate');
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.message || 'Session invalid');
    }
  }
);

export const login = createAsyncThunk(
  'auth/login',
  async ({ email, password }, { rejectWithValue }) => {
    try {
      const response = await axios.post('/auth/login', { email, password });
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.message || 'Login failed');
    }
  }
);

export const logout = createAsyncThunk(
  'auth/logout',
  async (_, { rejectWithValue }) => {
    try {
      await axios.get('/auth/logout');
      return true;
    } catch (error) {
      return rejectWithValue(error.response?.data?.message || 'Logout failed');
    }
  }
);

const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,
  },
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Validate Session
      .addCase(validateSession.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(validateSession.fulfilled, (state, action) => {
        state.isAuthenticated = true;
        state.user = action.payload;
        state.isLoading = false;
        state.error = null;
      })
      .addCase(validateSession.rejected, (state) => {
        state.isAuthenticated = false;
        state.user = null;
        state.isLoading = false;
      })
      
      // Login
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isAuthenticated = true;
        state.user = action.payload;
        state.isLoading = false;
        state.error = null;
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
      })
      
      // Logout
      .addCase(logout.fulfilled, (state) => {
        state.isAuthenticated = false;
        state.user = null;
        state.error = null;
      });
  },
});

export const { clearError } = authSlice.actions;
export default authSlice.reducer;
```

---

### 3. Vue 3 + Pinia (Composition API)

#### `src/stores/auth.js`
```js
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import axios from 'axios';

axios.defaults.withCredentials = true;
axios.defaults.baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null);
  const isLoading = ref(true);
  const error = ref(null);

  const isAuthenticated = computed(() => user.value !== null);

  const validateSession = async () => {
    try {
      const response = await axios.get('/auth/validate');
      if (response.data.message === 'success') {
        user.value = response.data.data;
      }
    } catch (err) {
      user.value = null;
    } finally {
      isLoading.value = false;
    }
  };

  const login = async (email, password) => {
    try {
      error.value = null;
      const response = await axios.post('/auth/login', { email, password });
      
      if (response.data.message === 'success') {
        user.value = response.data.data;
        return { success: true };
      }
      
      return { success: false, message: 'Login failed' };
    } catch (err) {
      const message = err.response?.data?.message || 'Login failed';
      error.value = message;
      return { success: false, message };
    }
  };

  const logout = async () => {
    try {
      await axios.get('/auth/logout');
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      user.value = null;
      error.value = null;
    }
  };

  return {
    user,
    isLoading,
    error,
    isAuthenticated,
    validateSession,
    login,
    logout,
  };
});
```

#### `src/router/index.js`
```js
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const routes = [
  {
    path: '/login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { requiresAuth: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  
  // Wait for initial session validation
  if (authStore.isLoading) {
    await authStore.validateSession();
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login');
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/dashboard');
  } else {
    next();
  }
});

export default router;
```

---

## Axios Configuration (All Frameworks)

### `src/api/axios.js`
```js
import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_URL,
  withCredentials: true, // CRITICAL: Send cookies with requests
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for global error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid - redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;
```

---

## Production Environment Setup

### 1. Backend Configuration (Go)

Ensure CORS is properly configured for your frontend domain:

```go
// In your main.go or middleware setup
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{"https://yourdomain.com", "http://localhost:3000"},
    AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
    AllowCredentials: true, // CRITICAL for cookies
}))
```

### 2. Cookie Configuration

**Development (localhost):**
```go
cookie := &http.Cookie{
    Name:     "Bearer",
    Value:    token,
    HttpOnly: true,
    Secure:   false,  // false for http://localhost
    Path:     "/",
    MaxAge:   86400,
    SameSite: http.SameSiteLaxMode, // Lax for local dev
}
```

**Production (HTTPS):**
```go
cookie := &http.Cookie{
    Name:     "Bearer",
    Value:    token,
    HttpOnly: true,
    Secure:   true,   // true for HTTPS
    Path:     "/",
    MaxAge:   86400,
    SameSite: http.SameSiteNoneMode, // None for cross-domain
    Domain:   ".yourdomain.com", // Optional: share across subdomains
}
```

### 3. Environment Variables

**Frontend (.env)**
```env
# Development
REACT_APP_API_URL=http://localhost:8080/api/v1

# Production
REACT_APP_API_URL=https://api.yourdomain.com/api/v1
```

**Backend (config)**
```env
# Development
JWT_SECRET=your-secret-key-development
FRONTEND_URL=http://localhost:3000
COOKIE_SECURE=false
COOKIE_SAMESITE=Lax

# Production
JWT_SECRET=your-very-secure-secret-key-production
FRONTEND_URL=https://yourdomain.com
COOKIE_SECURE=true
COOKIE_SAMESITE=None
```

---

## Security Best Practices

### ✅ Implemented in Your Backend
- ✓ HTTP-only cookies (XSS protection)
- ✓ Secure flag (HTTPS only in production)
- ✓ SameSite=None (CSRF protection for cross-origin)
- ✓ Short token expiry (24 hours)

### 🔒 Additional Recommendations

1. **Add CSRF Protection** (if needed for state-changing requests)
2. **Rate Limiting** on auth endpoints
3. **Implement Refresh Token** (optional, but recommended for long sessions)
4. **Token Rotation** - Issue new token on each request
5. **Logout on All Devices** - Track sessions in Redis/DB

---

## Testing Authentication

### Test Session Validation
```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  -c cookies.txt

# 2. Validate Session
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -b cookies.txt

# 3. Access Protected Route
curl -X GET http://localhost:8080/api/v1/books/getall \
  -b cookies.txt

# 4. Logout
curl -X GET http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

---

## Common Issues & Solutions

### Issue: Cookies not sent with requests
**Solution:** Ensure `withCredentials: true` in axios config

### Issue: CORS errors in browser
**Solution:** Backend must set `AllowCredentials: true` and specific origin (not `*`)

### Issue: Cookies work in Postman but not browser
**Solution:** Check SameSite attribute and ensure HTTPS in production

### Issue: Session lost on page refresh
**Solution:** Call `validateSession()` in app initialization

---

## Deployment Checklist

- [ ] Set `Secure: true` for cookies in production
- [ ] Configure CORS with exact frontend domain
- [ ] Use HTTPS for both frontend and backend
- [ ] Set strong JWT_SECRET (minimum 32 characters)
- [ ] Add rate limiting on auth endpoints
- [ ] Implement logging for failed auth attempts
- [ ] Set proper token expiry time
- [ ] Test auth flow in production environment
- [ ] Add monitoring for 401 errors

---

## Need Refresh Tokens?

If you want to implement refresh tokens for longer sessions without storing them in cookies, I can provide that implementation as well. Let me know!
