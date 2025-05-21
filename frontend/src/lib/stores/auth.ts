import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

const API_BASE_URL = 'http://localhost:9000/v1'; 

export interface User {
  id: string; 
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  isLoading: boolean;
  error: string | null;
}

const initialAuthState: AuthState = {
  isAuthenticated: false,
  user: null,
  isLoading: true, 
  error: null,
};

export const authStore: Writable<AuthState> = writable(initialAuthState);

// Server-side authentication check (for use in server load functions)
export async function checkAuth(cookies: import('@sveltejs/kit').Cookies) {
  if (!cookies) {
    console.warn('checkAuth called with undefined cookies');
    return null;
  } 
  
  const token = cookies.get('token');
  if (!token) return null;
  
  try {
    const res = await fetch(`${API_BASE_URL}/users/me`, {
      headers: {
        cookie: `token=${token}`
      },
      credentials: 'include'
    });
    
    if (!res.ok) return null;
    
    const user = await res.json();
    return user;
  } catch (error) {
    console.error('Auth check failed:', error);
    return null;
  }
}

// Client-side authentication check (for use in browser)
export async function checkAuthClient(): Promise<User | null> {
  if (!browser) return null;
  
  try {
    // Set loading state
    authStore.update(state => ({ ...state, isLoading: true }));
    
    const res = await fetch(`${API_BASE_URL}/users/me`, {
      credentials: 'include' // This will include the httpOnly cookie automatically
    });
    
    if (!res.ok) {
      // Not authenticated
      authStore.update(state => ({
        ...state,
        isAuthenticated: false,
        user: null,
        isLoading: false,
        error: null
      }));
      return null;
    }
    
    const user = await res.json();
    
    // Update store with authenticated user
    authStore.update(state => ({
      ...state,
      isAuthenticated: true,
      user: user,
      isLoading: false,
      error: null
    }));
    
    return user;
  } catch (error) {
    console.error('Client auth check failed:', error);
    authStore.update(state => ({
      ...state,
      isAuthenticated: false,
      user: null,
      isLoading: false,
      error: 'Failed to check authentication status'
    }));
    return null;
  }
}

// Login function
export async function login(email: string, password: string): Promise<{ success: boolean; error?: string }> {
  try {
    authStore.update(state => ({ ...state, isLoading: true, error: null }));
    
    const res = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    });
    
    if (!res.ok) {
      const errorData = await res.json().catch(() => ({ message: 'Login failed' }));
      const errorMessage = errorData.message || 'Login failed';
      
      authStore.update(state => ({
        ...state,
        isLoading: false,
        error: errorMessage
      }));
      
      return { success: false, error: errorMessage };
    }
    
    // After successful login, check auth status to get user data
    await checkAuthClient();
    
    return { success: true };
  } catch (error) {
    const errorMessage = 'Network error during login';
    authStore.update(state => ({
      ...state,
      isLoading: false,
      error: errorMessage
    }));
    return { success: false, error: errorMessage };
  }
}

// Logout function
export async function logout(): Promise<void> {
  try {
    await fetch(`${API_BASE_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include'
    });
  } catch (error) {
    console.error('Logout request failed:', error);
  } finally {
    // Always clear the store regardless of API call success
    authStore.update(state => ({
      ...state,
      isAuthenticated: false,
      user: null,
      isLoading: false,
      error: null
    }));
  }
}

// Initialize auth check when in browser
if (browser) {
  checkAuthClient();
}
