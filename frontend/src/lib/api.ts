export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  meta?: {
    request_id?: string
    timestamp?: string
  }
  error?: {
    code: string
    message: string
    request_id?: string
    details?: any
  }
}

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

class ApiClient {
  private getAccessToken(): string | null {
    try {
      const authData = localStorage.getItem('recess_auth')
      if (authData) {
        const parsed = JSON.parse(authData)
        return parsed.state?.tokens?.access_token || null
      }
    } catch {
      return null
    }
    return null
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = this.getAccessToken()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const url = endpoint.startsWith('http') ? endpoint : `${API_BASE}${endpoint}`

    const response = await fetch(url, {
      ...options,
      headers,
    })

    const payload: ApiResponse<T> = await response.json().catch(() => ({
      success: false,
      error: { code: 'NETWORK_ERROR', message: 'Failed to parse response' },
    }))

    if (!response.ok || !payload.success) {
      const message = payload.error?.message || `Request failed with status ${response.status}`
      const code = payload.error?.code || 'UNKNOWN_ERROR'
      
      const error: any = new Error(message)
      error.code = code
      error.status = response.status
      error.details = payload.error?.details
      throw error
    }

    return payload.data as T
  }

  get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  post<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  put<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

export const api = new ApiClient()
