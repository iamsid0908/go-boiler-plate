# Next.js Authentication Implementation with HTTP-Only Cookies

This guide covers both **App Router** (Next.js 13+) and **Pages Router** implementations.

---

## App Router Implementation (Next.js 13+)

### 1. Auth Context Provider

#### `app/providers/auth-provider.tsx`
```tsx
'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  language: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; message?: string }>;
  logout: () => Promise<void>;
  validateSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  const isAuthenticated = user !== null;

  useEffect(() => {
    validateSession();
  }, []);

  const validateSession = async () => {
    try {
      const response = await fetch(`${API_URL}/auth/validate`, {
        method: 'GET',
        credentials: 'include', // CRITICAL: Include cookies
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        if (data.message === 'success') {
          setUser(data.data);
        }
      } else {
        setUser(null);
      }
    } catch (error) {
      console.error('Session validation failed:', error);
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, password: string) => {
    try {
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (response.ok && data.message === 'success') {
        setUser(data.data);
        return { success: true };
      }

      return { success: false, message: data.message || 'Login failed' };
    } catch (error) {
      return { success: false, message: 'Network error. Please try again.' };
    }
  };

  const logout = async () => {
    try {
      await fetch(`${API_URL}/auth/logout`, {
        method: 'GET',
        credentials: 'include',
      });
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      router.push('/login');
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated,
        isLoading,
        login,
        logout,
        validateSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
```

### 2. Layout with Provider

#### `app/layout.tsx`
```tsx
import { AuthProvider } from './providers/auth-provider';
import './globals.css';

export const metadata = {
  title: 'Book Finder',
  description: 'Your book management system',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
```

### 3. Middleware for Route Protection

#### `middleware.ts` (root level)
```ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Public routes that don't require authentication
  const publicRoutes = ['/login', '/register', '/verify-otp'];
  const isPublicRoute = publicRoutes.some(route => pathname.startsWith(route));

  if (isPublicRoute) {
    return NextResponse.next();
  }

  // Check authentication by calling validate endpoint
  try {
    const cookie = request.cookies.get('Bearer');
    
    if (!cookie) {
      return NextResponse.redirect(new URL('/login', request.url));
    }

    const response = await fetch(`${API_URL}/auth/validate`, {
      method: 'GET',
      headers: {
        'Cookie': `Bearer=${cookie.value}`,
      },
    });

    if (!response.ok) {
      return NextResponse.redirect(new URL('/login', request.url));
    }

    return NextResponse.next();
  } catch (error) {
    return NextResponse.redirect(new URL('/login', request.url));
  }
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder
     */
    '/((?!_next/static|_next/image|favicon.ico|public).*)',
  ],
};
```

### 4. Login Page

#### `app/login/page.tsx`
```tsx
'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../providers/auth-provider';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const { login } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    const result = await login(email, password);

    if (result.success) {
      router.push('/dashboard');
    } else {
      setError(result.message || 'Login failed');
    }

    setIsLoading(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow-md">
        <div>
          <h2 className="text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Email address"
              />
            </div>
            <div>
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Password"
              />
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
            >
              {isLoading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
```

### 5. Protected Dashboard Page

#### `app/dashboard/page.tsx`
```tsx
'use client';

import { useAuth } from '../providers/auth-provider';
import { useRouter } from 'next/navigation';

export default function DashboardPage() {
  const { user, logout, isLoading } = useAuth();
  const router = useRouter();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  if (!user) {
    router.push('/login');
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold">Book Finder</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">
                Welcome, {user.name}
              </span>
              <button
                onClick={logout}
                className="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-2xl font-bold mb-4">Dashboard</h2>
          <div className="space-y-2">
            <p><strong>Email:</strong> {user.email}</p>
            <p><strong>Role:</strong> {user.role}</p>
            <p><strong>Language:</strong> {user.language}</p>
          </div>
        </div>
      </main>
    </div>
  );
}
```

### 6. API Helper with Fetch

#### `lib/api.ts`
```ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: any;
  headers?: Record<string, string>;
}

export async function api<T = any>(
  endpoint: string,
  options: ApiOptions = {}
): Promise<{ data?: T; error?: string; success: boolean }> {
  const { method = 'GET', body, headers = {} } = options;

  try {
    const response = await fetch(`${API_URL}${endpoint}`, {
      method,
      credentials: 'include', // CRITICAL: Send cookies
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        error: data.message || 'Request failed',
      };
    }

    return {
      success: true,
      data: data.data,
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Network error',
    };
  }
}

// Usage examples:
export const bookApi = {
  getAll: () => api('/books/getall'),
  insert: (book: any) => api('/books/insert', { method: 'POST', body: book }),
  recommend: (data: any) => api('/books/recommendation/books', { method: 'POST', body: data }),
};

export const cartApi = {
  get: () => api('/cart/get-cart'),
  add: (bookId: string) => api('/cart/insert', { method: 'POST', body: { book_id: bookId } }),
  remove: (bookId: string) => api('/cart/cart-remove', { method: 'DELETE', body: { book_id: bookId } }),
  getSize: () => api('/cart/cart-size'),
};
```

---

## Server-Side Authentication (App Router)

For true server-side authentication, use Server Components and Server Actions:

### `app/actions/auth.ts` (Server Actions)
```ts
'use server';

import { cookies } from 'next/headers';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export async function serverLogin(email: string, password: string) {
  try {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();

    // Extract cookie from response
    const setCookie = response.headers.get('set-cookie');
    if (setCookie && data.message === 'success') {
      // Forward the cookie to the client
      const cookieStore = cookies();
      // Parse and set the cookie
      // Note: In production, you might need more sophisticated cookie parsing
      cookieStore.set('Bearer', data.data.token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        maxAge: 86400,
      });

      return { success: true, data: data.data };
    }

    return { success: false, message: data.message };
  } catch (error) {
    return { success: false, message: 'Login failed' };
  }
}

export async function serverValidateSession() {
  const cookieStore = cookies();
  const token = cookieStore.get('Bearer');

  if (!token) {
    return { isAuthenticated: false };
  }

  try {
    const response = await fetch(`${API_URL}/auth/validate`, {
      headers: {
        Cookie: `Bearer=${token.value}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      return { isAuthenticated: true, user: data.data };
    }

    return { isAuthenticated: false };
  } catch (error) {
    return { isAuthenticated: false };
  }
}
```

---

## Pages Router Implementation (Next.js 12 and earlier)

### `pages/_app.tsx`
```tsx
import type { AppProps } from 'next/app';
import { AuthProvider } from '../context/AuthContext';
import '../styles/globals.css';

export default function App({ Component, pageProps }: AppProps) {
  return (
    <AuthProvider>
      <Component {...pageProps} />
    </AuthProvider>
  );
}
```

### `context/AuthContext.tsx`
```tsx
// Use the same AuthProvider code from App Router above
// Just import from 'next/router' instead of 'next/navigation'

import { useRouter } from 'next/router';

// ... rest of the code is the same
```

### `pages/login.tsx`
```tsx
import { useState } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '../context/AuthContext';

export default function Login() {
  // Same implementation as App Router
  // Just use useRouter from 'next/router'
}
```

---

## Environment Configuration

### `.env.local`
```env
# Development
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# Production (add to Vercel/deployment platform)
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
```

---

## Deployment Configuration

### Vercel Deployment

**vercel.json**
```json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        }
      ]
    }
  ]
}
```

### Backend CORS for Production
```go
// Update your CORS config
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{
        "https://yourdomain.com",
        "https://www.yourdomain.com",
        "http://localhost:3000", // for local dev
    },
    AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
    AllowCredentials: true,
    ExposeHeaders:    []string{"Set-Cookie"},
}))
```

---

## Testing

```bash
# Install dependencies
npm install
# or
yarn install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

---

## Common Next.js Issues

### Issue: Cookies not persisting after deployment
**Solution:** Ensure your API and frontend are on the same root domain or use proper CORS headers

### Issue: 401 errors on client-side navigation
**Solution:** Add credentials: 'include' to all fetch calls

### Issue: Middleware runs too often
**Solution:** Use proper matcher patterns in middleware.ts

### Issue: Session lost on page refresh in production
**Solution:** Verify cookie domain and SameSite settings match your deployment

---

## Security Enhancements

1. **Add CSRF tokens** for state-changing operations
2. **Implement rate limiting** on auth routes
3. **Use environment variables** for all sensitive config
4. **Enable HTTP Strict Transport Security** (HSTS)
5. **Implement proper error logging**

This implementation is production-ready and follows Next.js best practices!
