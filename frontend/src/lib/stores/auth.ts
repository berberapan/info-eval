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

export async function checkAuthClient(): Promise<User | null> {
  if (!browser) return null;
  
  try {
    authStore.update(state => ({ ...state, isLoading: true }));
    const res = await fetch(`${API_BASE_URL}/users/me`, {
      credentials: 'include' 
    });
    if (!res.ok) {
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

if (browser) {
  checkAuthClient();
}
