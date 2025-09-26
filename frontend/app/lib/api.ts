import { getConnectionInfo } from './storage'

export interface ApiResponse<T = any> {
  data: T
  message?: string
  error?: string
}

export interface Task {
  id: string
  instruction: string
  branch: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  agent: string
  createdAt: string
  updatedAt: string
}

export interface Patch {
  file: string
  patch: string
  hunks: Array<{
    startOld: number
    lenOld: number
    startNew: number
    lenNew: number
    lines: string[]
  }>
}

export interface PatchSelection {
  file: string
  hunks: number[]
}

class ApiClient {
  private apiBase: string = ''

  constructor() {
    this.initializeApiBase()
  }

  private async initializeApiBase(): Promise<void> {
    const connectionInfo = await getConnectionInfo()
    if (connectionInfo?.apiBase) {
      this.apiBase = connectionInfo.apiBase
    }
  }

  private async getHeaders(): Promise<Record<string, string>> {
    const connectionInfo = await getConnectionInfo()
    const token = connectionInfo?.token

    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    try {
      const url = `${this.apiBase}${endpoint}`
      const headers = await this.getHeaders()

      const response = await fetch(url, {
        ...options,
        headers: {
          ...headers,
          ...options.headers,
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Request failed')
    }
  }

  async createSession(repo: string, label: string, via: string): Promise<any> {
    return this.request('/api/session', {
      method: 'POST',
      body: JSON.stringify({ repo, label, via }),
    })
  }

  async getSession(id: string): Promise<any> {
    return this.request(`/api/session/${id}`)
  }

  async createTask(instruction: string, branch: string, context: any, agent: string): Promise<Task> {
    return this.request('/api/tasks', {
      method: 'POST',
      body: JSON.stringify({ instruction, branch, context, agent }),
    })
  }

  async getTask(id: string): Promise<Task> {
    return this.request(`/api/tasks/${id}`)
  }

  async getTaskPatches(id: string): Promise<{ patches: Patch[] }> {
    return this.request(`/api/tasks/${id}/patches`)
  }

  async applyTaskPatches(id: string, selections: PatchSelection[], commitMessage: string): Promise<any> {
    return this.request(`/api/tasks/${id}/apply`, {
      method: 'POST',
      body: JSON.stringify({ select: selections, commitMessage }),
    })
  }

  async runCommand(cmd: string, cwd: string, timeoutMs: number): Promise<any> {
    return this.request('/api/cmd', {
      method: 'POST',
      body: JSON.stringify({ cmd, cwd, timeoutMs }),
    })
  }
}

export const apiClient = new ApiClient()
