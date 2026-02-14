# Quick Start Guide - HTTP-Only Cookie Authentication

## What I Fixed in Your Backend

### 🐛 Critical Bug Fixed
- **Logout cookie mismatch**: Changed from `accessToken` to `Bearer` cookie name
- **Token exposure**: Removed token from login response body (security improvement)

### ✅ New Features Added
1. **Session Validation Endpoint**: `GET /auth/validate`
   - Frontend can check if user is authenticated
   - Returns user data without exposing token

2. **Enhanced Logout**: Properly clears the Bearer cookie with correct settings

---

## Quick Implementation Guide

### For React/Next.js Frontend

**Step 1: Configure Axios**
```js
import axios from 'axios';

axios.defaults.baseURL = 'http://localhost:8080/api/v1';
axios.defaults.withCredentials = true; // CRITICAL!
```

**Step 2: Create Auth Context (see FRONTEND_AUTH_GUIDE.md)**

**Step 3: Protect Routes**
```jsx
<ProtectedRoute>
  <Dashboard />
</ProtectedRoute>
```

---

## API Endpoints

### Public Routes
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/verify-otp` - Verify email OTP
- `POST /api/v1/auth/login` - Login (sets HTTP-only cookie)
- `POST /api/v1/auth/resend-otp` - Resend OTP

### Protected Routes (Require Cookie)
- `GET /api/v1/auth/validate` - Check session validity
- `GET /api/v1/auth/logout` - Logout (clears cookie)
- `GET /api/v1/user/get-user` - Get user info
- `GET /api/v1/books/*` - All book endpoints
- `GET /api/v1/cart/*` - All cart endpoints
- All other routes with middleware.JWTVerify()

---

## Testing Your Implementation

### 1. Test Backend Cookie Flow
```bash
# Login and save cookie
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your@email.com","password":"yourpassword"}' \
  -c cookies.txt -v

# Validate session with cookie
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -b cookies.txt

# Access protected route
curl -X GET http://localhost:8080/api/v1/books/getall \
  -b cookies.txt

# Logout
curl -X GET http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

### 2. Test Frontend Integration
```js
// Login
const login = async (email, password) => {
  const response = await axios.post('/auth/login', { email, password });
  console.log('User data:', response.data.data);
  // Cookie is automatically stored by browser
};

// Validate (automatic cookie send)
const validate = async () => {
  const response = await axios.get('/auth/validate');
  console.log('Session valid:', response.data);
};

// Logout
const logout = async () => {
  await axios.get('/auth/logout');
  // Cookie is automatically cleared
};
```

---

## Environment Setup

### Development
```env
# Backend .env
JWT_SECRET=development-secret-key-change-in-production
COOKIE_SECURE=false
COOKIE_SAMESITE=Lax
FRONTEND_URL=http://localhost:3000

# Frontend .env
REACT_APP_API_URL=http://localhost:8080/api/v1
```

### Production
```env
# Backend .env
JWT_SECRET=super-secure-production-key-min-64-chars
COOKIE_SECURE=true
COOKIE_SAMESITE=None
COOKIE_DOMAIN=.yourdomain.com
FRONTEND_URL=https://yourdomain.com

# Frontend .env
REACT_APP_API_URL=https://api.yourdomain.com/api/v1
```

---

## How HTTP-Only Cookies Work

### Login Flow
```
1. User submits credentials
   Browser → POST /auth/login → Backend

2. Backend validates & creates JWT
   Backend → Set-Cookie: Bearer=<token>; HttpOnly

3. Browser stores cookie automatically
   Cookie Storage (inaccessible to JavaScript)

4. Frontend updates state with user data
   Response.data.data → User State
```

### Authenticated Requests
```
1. User makes request
   Browser → GET /books/getall

2. Browser automatically sends cookie
   Headers: Cookie: Bearer=<token>

3. Backend middleware validates token
   Extract from cookie → Verify JWT → Allow access

4. Response returned
   User data → Frontend
```

### Logout Flow
```
1. User clicks logout
   Browser → GET /auth/logout

2. Backend clears cookie
   Set-Cookie: Bearer=; MaxAge=-1

3. Frontend clears state
   User state → null
```

---

## Security Advantages

### ✅ What You Get
- **XSS Protection**: JavaScript cannot access HTTP-only cookies
- **Automatic Sending**: Browser handles cookie management
- **Secure**: Sent only over HTTPS (in production)
- **CSRF Protection**: SameSite attribute
- **Simple**: No manual token storage/retrieval

### ❌ What to Avoid
- Don't store tokens in localStorage
- Don't send tokens in response body (already fixed)
- Don't use SameSite=None without HTTPS
- Don't use wildcard (*) in CORS origins with credentials

---

## Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Cookie not sent | Add `withCredentials: true` to axios |
| CORS error | Set `AllowCredentials: true` in backend |
| Cookie not set | Check Secure flag matches protocol (HTTP/HTTPS) |
| Session lost on refresh | Call `/auth/validate` on app mount |
| 401 on all requests | Verify cookie domain matches |

---

## Complete File Structure

```
Your Project/
├── Backend (Go)
│   ├── handler/
│   │   └── auth.go (✅ Fixed logout, added validate)
│   ├── route/
│   │   └── v1.go (✅ Added validate route)
│   ├── middleware/
│   │   └── jwt.go (Already working)
│   ├── service/
│   │   └── auth.go (Extracts JWT from cookie)
│   └── GUIDES/
│       ├── FRONTEND_AUTH_GUIDE.md (React/Vue examples)
│       ├── NEXTJS_AUTH_IMPLEMENTATION.md (Next.js specific)
│       └── PRODUCTION_SECURITY_GUIDE.md (Advanced security)
│
└── Frontend (React/Next.js)
    ├── context/
    │   └── AuthContext.jsx (State management)
    ├── components/
    │   └── ProtectedRoute.jsx (Route guard)
    ├── api/
    │   └── axios.js (Configured with credentials)
    └── pages/
        ├── Login.jsx
        ├── Dashboard.jsx
        └── ...
```

---

## Next Steps

1. **Choose Your Frontend Framework**
   - React: Use `FRONTEND_AUTH_GUIDE.md`
   - Next.js: Use `NEXTJS_AUTH_IMPLEMENTATION.md`
   - Vue: Check Vue section in guide

2. **Implement Auth Context**
   - Copy the auth context code
   - Configure axios with `withCredentials: true`

3. **Test Locally**
   - Start backend: `go run main.go`
   - Start frontend: `npm run dev`
   - Test login → validate → protected routes → logout

4. **Add Production Features** (Optional)
   - Rate limiting (see PRODUCTION_SECURITY_GUIDE.md)
   - Account lockout after failed attempts
   - Security headers middleware
   - Logging and monitoring

5. **Deploy**
   - Set environment variables
   - Enable HTTPS
   - Update CORS settings
   - Test in production environment

---

## Documentation Files

📄 **FRONTEND_AUTH_GUIDE.md** - Complete frontend implementation examples
   - React + Context API ⭐ Recommended for most apps
   - React + Redux Toolkit (for large apps)
   - Vue 3 + Pinia
   - Axios configuration
   - Testing guide

📄 **NEXTJS_AUTH_IMPLEMENTATION.md** - Next.js specific guide
   - App Router (Next.js 13+)
   - Pages Router (Next.js 12)
   - Server Components
   - Middleware for route protection
   - Deployment on Vercel

📄 **PRODUCTION_SECURITY_GUIDE.md** - Advanced security features
   - Environment-based cookie config
   - Rate limiting
   - Account lockout
   - Security headers
   - Logging & monitoring
   - Refresh token pattern (optional)

---

## Support & Troubleshooting

If you encounter issues:

1. Check browser DevTools → Network tab → Cookie column
2. Verify CORS configuration matches your frontend URL
3. Confirm `withCredentials: true` in all API calls
4. Test with curl first (see testing section above)
5. Check backend logs for JWT errors

---

## Quick Reference: Key Changes Made

### Backend Changes
```diff
# handler/auth.go
- cookie.Name = "accessToken"  // WRONG
+ cookie.Name = "Bearer"       // FIXED

- Data: data (includes token)  // INSECURE
+ Data: {...} (no token)       // SECURE

+ ValidateSession endpoint     // NEW
```

### What You Need to Do
1. Create auth context in frontend
2. Set `withCredentials: true` in axios
3. Call `/auth/validate` on app load
4. Protect routes with auth check

That's it! Your backend is ready for production-level authentication. 🚀
