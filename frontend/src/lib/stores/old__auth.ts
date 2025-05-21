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
  if(!cookies) {
      console.warn('checkAuth called with undefined cookies');
      return null;
  } 
  const token = cookies.get('token');
  if (!token) return null;

  const res = await fetch(`${API_BASE_URL}/users/me`, {
    headers: {
      cookie: `token=${token}`
    },
    credentials: 'include'
  });

  if (!res.ok) return null;

  const user = await res.json();
  return user;
}

if (browser) {
  checkAuth();
}

