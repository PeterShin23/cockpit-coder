import * as SecureStore from 'expo-secure-store'

interface Storage {
  getItem: (key: string) => Promise<string | null>
  setItem: (key: string, value: string) => Promise<void>
  removeItem: (key: string) => Promise<void>
}

export function useSecureStore(): Storage {
  return {
    getItem: async (key: string) => {
      try {
        return await SecureStore.getItemAsync(key)
      } catch {
        return null
      }
    },
    setItem: async (key: string, value: string) => {
      await SecureStore.setItemAsync(key, value)
    },
    removeItem: async (key: string) => {
      await SecureStore.deleteItemAsync(key)
    },
  }
}

export async function saveConnectionInfo(info: any): Promise<void> {
  const json = JSON.stringify(info)
  await SecureStore.setItemAsync('connectionInfo', json)
}

export async function getConnectionInfo(): Promise<any | null> {
  try {
    const json = await SecureStore.getItemAsync('connectionInfo')
    return json ? JSON.parse(json) : null
  } catch {
    return null
  }
}

export async function clearConnectionInfo(): Promise<void> {
  await SecureStore.deleteItemAsync('connectionInfo')
}
