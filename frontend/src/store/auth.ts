import { defineStore } from 'pinia'
import axios from 'axios'

interface UserState {
  token: string | null
  expiresAt: number | null
}

export const useAuthStore = defineStore('auth', {
  state: (): UserState => ({
    token: localStorage.getItem('token'),
    expiresAt: Number(localStorage.getItem('expiresAt')) || null
  }),
  getters: {
    isAuthenticated: (state) => {
      if (!state.token) return false
      if (!state.expiresAt) return false
      return Date.now() < state.expiresAt
    }
  },
  actions: {
    async login(username: string, password: string) {
      const response = await axios.post('/api/auth/login', { username, password })
      const { token, expires_in } = response.data
      
      this.token = token
      this.expiresAt = Date.now() + expires_in * 1000
      
      localStorage.setItem('token', token)
      localStorage.setItem('expiresAt', this.expiresAt.toString())
      
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
    },
    logout() {
      this.token = null
      this.expiresAt = null
      localStorage.removeItem('token')
      localStorage.removeItem('expiresAt')
      delete axios.defaults.headers.common['Authorization']
    },
    init() {
      if (this.token) {
        axios.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
      }
    }
  }
})
