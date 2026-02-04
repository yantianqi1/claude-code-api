import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authApi } from '@/api/client';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (apiKey: string) => Promise<{ success: boolean; message: string }>;
  logout: () => void;
  verify: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if already authenticated by verifying with server
    const checkAuth = async () => {
      try {
        const result = await authApi.verify();
        setIsAuthenticated(result.data.authenticated);
      } catch {
        setIsAuthenticated(false);
      } finally {
        setIsLoading(false);
      }
    };
    checkAuth();
  }, []);

  const login = async (apiKey: string) => {
    try {
      const response = await authApi.login(apiKey);
      if (response.data.success) {
        setIsAuthenticated(true);
      }
      return response.data;
    } catch {
      return {
        success: false,
        message: '登录失败，请检查网络连接',
      };
    }
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } finally {
      setIsAuthenticated(false);
    }
  };

  const verify = async () => {
    try {
      const result = await authApi.verify();
      setIsAuthenticated(result.data.authenticated);
      return result.data.authenticated;
    } catch {
      setIsAuthenticated(false);
      return false;
    }
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading, login, logout, verify }}>
      {children}
    </AuthContext.Provider>
  );
};
