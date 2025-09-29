import React, { useState, useEffect } from 'react'
import { View, Text, StyleSheet, Alert, ActivityIndicator } from 'react-native'
import { Camera, CameraView, BarcodeScanningResult } from 'expo-camera'
import { Button } from '../components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { useSecureStore } from '../lib/storage'
import { apiClient } from '../lib/api'

interface ConnectionInfo {
  sessionId: string
  token: string
  ws: {
    pty?: string
    events?: string
    host?: string
    client?: string
  }
  apiBase: string
  expiresAt: string
}

export function QRConnectScreen({ navigation }: { navigation?: any }) {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null)
  const [scanned, setScanned] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [connectionInfo, setConnectionInfo] = useState<ConnectionInfo | null>(null)

  const { getItem, setItem } = useSecureStore()

  useEffect(() => {
    const getCameraPermissions = async () => {
      const { status } = await Camera.requestCameraPermissionsAsync()
      setHasPermission(status === 'granted')
    }

    getCameraPermissions()
  }, [])

  const handleBarCodeScanned = async ({ data }: BarcodeScanningResult) => {
    if (scanned || connecting) return

    setScanned(true)
    setConnecting(true)

    try {
      const connectionData = JSON.parse(data)
      
      // Validate connection data structure
      if (!connectionData.sessionId || !connectionData.token || !connectionData.ws) {
        throw new Error('Invalid QR code format')
      }

      // Check if we have either local or relay WebSocket URLs
      const hasLocalWS = connectionData.ws.pty && connectionData.ws.events;
      const hasRelayWS = connectionData.ws.host && connectionData.ws.client;
      
      if (!hasLocalWS && !hasRelayWS) {
        throw new Error('Invalid WebSocket configuration in QR code')
      }

      // Store connection info
      await setItem('connectionInfo', JSON.stringify(connectionData))
      
      // Validate connection by opening appropriate WebSocket
      let ws: WebSocket;
      if (hasLocalWS) {
        // Local mode - connect to events WebSocket
        ws = new WebSocket(`${connectionData.ws.events}?sessionId=${connectionData.sessionId}&token=${connectionData.token}`)
      } else {
        // Relay mode - connect to client WebSocket
        ws = new WebSocket(`${connectionData.ws.client}?sessionId=${connectionData.sessionId}&token=${connectionData.token}`)
      }
      
      ws.onopen = () => {
        setConnectionInfo(connectionData)
        ws.close()
      }
      
      ws.onerror = () => {
        throw new Error('Failed to connect to server')
      }
      
      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          if (message.type === 'error') {
            throw new Error(message.message || 'Connection failed')
          }
        } catch (e) {
          ws.close()
        }
      }
      
      // Set timeout for validation
      setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN) {
          ws.close()
          throw new Error('Connection validation timeout')
        }
      }, 5000)

    } catch (error) {
      Alert.alert('Connection Error', error instanceof Error ? error.message : 'Failed to connect')
      setScanned(false)
      setConnecting(false)
    }
  }

  const handleConnect = () => {
    if (connectionInfo) {
      // Navigate to chat screen
      navigation?.navigate('chat')
    }
  }

  if (hasPermission === null) {
    return <View style={styles.container}><ActivityIndicator size="large" /></View>
  }

  if (hasPermission === false) {
    return (
      <View style={styles.container}>
        <Card>
          <CardHeader>
            <CardTitle>No Camera Access</CardTitle>
          </CardHeader>
          <CardContent>
            <Text style={styles.text}>Camera access is required to scan QR codes.</Text>
            <Button title="Grant Permission" onPress={() => {}} />
          </CardContent>
        </Card>
      </View>
    )
  }

  return (
    <View style={styles.container}>
      <Card>
        <CardHeader>
          <CardTitle>Connect to Coding Agent</CardTitle>
        </CardHeader>
        <CardContent>
          <Text style={styles.text}>
            Scan the QR code from your coding agent to connect
          </Text>
          
          {connecting ? (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="large" />
              <Text style={styles.loadingText}>Validating connection...</Text>
            </View>
          ) : (
            <CameraView
              style={styles.camera}
              onBarcodeScanned={scanned ? undefined : handleBarCodeScanned}
              barcodeScannerSettings={{
                barcodeTypes: ['qr'],
              }}
            />
          )}
          
          {connectionInfo && (
            <View style={styles.successContainer}>
              <Text style={styles.successText}>Connected successfully!</Text>
              <Button title="Open Terminal" onPress={handleConnect} />
            </View>
          )}
          
          {scanned && !connecting && !connectionInfo && (
            <Button 
              title="Rescan" 
              onPress={() => setScanned(false)} 
              variant="outline"
            />
          )}
        </CardContent>
      </Card>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  text: {
    fontSize: 16,
    marginBottom: 16,
    textAlign: 'center',
  },
  camera: {
    width: '100%',
    height: 300,
    borderRadius: 8,
    overflow: 'hidden',
  },
  loadingContainer: {
    padding: 20,
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
  },
  successContainer: {
    marginTop: 16,
    alignItems: 'center',
  },
  successText: {
    fontSize: 16,
    color: '#059669',
    marginBottom: 16,
  },
})
