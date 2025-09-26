import React, { useState, useEffect, useRef } from 'react'
import { View, StyleSheet, Alert, BackHandler } from 'react-native'
import { WebView } from 'react-native-webview'
import { Button } from '../components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { getConnectionInfo } from '../lib/storage'
import { createWebSocketClient, WSEvent } from '../lib/wsClient'

interface TerminalScreenProps {
  navigation?: any
}

export function TerminalScreen({ navigation }: TerminalScreenProps) {
  const webViewRef = useRef<WebView>(null)
  const [connectionInfo, setConnectionInfo] = useState<any>(null)
  const [eventsClient, setEventsClient] = React.useState<any>(null)
  const [terminalUrl, setTerminalUrl] = useState<string>('')

  useEffect(() => {
    const loadConnectionInfo = async () => {
      try {
        const info = await getConnectionInfo()
        if (info) {
          setConnectionInfo(info)
          setTerminalUrl(`file:///assets/terminal/index.html?sessionId=${info.sessionId}&token=${info.token}`)
          
          // Create events WebSocket client
          const client = createWebSocketClient(info.ws.events)
          await client.connect()
          
          client.on('status', (event: WSEvent) => {
            console.log('Status event:', event)
          })
          
          client.on('patch', (event: WSEvent) => {
            console.log('Patch event:', event)
          })
          
          client.on('error', (event: WSEvent) => {
            Alert.alert('Error', event.message || 'Unknown error')
          })
          
          setEventsClient(client)
        } else {
          Alert.alert('No Connection', 'Please scan a QR code to connect first.')
          navigation?.navigate('chat')
        }
      } catch (error) {
        console.error('Error loading connection info:', error)
        Alert.alert('Connection Error', 'Failed to load connection information.')
      }
    }

    loadConnectionInfo()

    // Handle back button
    const backHandler = () => {
      if (eventsClient) {
        eventsClient.disconnect()
      }
      return false
    }

    BackHandler.addEventListener('hardwareBackPress', backHandler)

    return () => {
      BackHandler.removeEventListener('hardwareBackPress', backHandler)
      if (eventsClient) {
        eventsClient.disconnect()
      }
    }
  }, [navigation])

  const handleResize = () => {
    // Handle terminal resize based on orientation
    // This would typically be implemented with device orientation API
    console.log('Terminal resize requested')
  }

  const handleCancel = () => {
    if (eventsClient) {
      eventsClient.disconnect()
    }
    navigation?.navigate('chat')
  }

  const handleWebViewMessage = (event: any) => {
    try {
      const message = JSON.parse(event.nativeEvent.data)
      console.log('WebView message:', message)
      
      // Handle messages from WebView if needed
      if (message.type === 'resize' && eventsClient) {
        eventsClient.send(message)
      }
    } catch (error) {
      console.error('Error parsing WebView message:', error)
    }
  }

  if (!connectionInfo) {
    return (
      <View style={styles.container}>
        <Card>
          <CardContent>
            <View style={styles.loadingContainer}>
              <CardTitle>Loading connection...</CardTitle>
            </View>
          </CardContent>
        </Card>
      </View>
    )
  }

  return (
    <View style={styles.container}>
      <Card>
        <CardHeader>
          <CardTitle>Terminal</CardTitle>
        </CardHeader>
        <CardContent>
          <View style={styles.terminalContainer}>
            {terminalUrl ? (
              <WebView
                ref={webViewRef}
                source={{ uri: terminalUrl }}
                style={styles.terminal}
                onMessage={handleWebViewMessage}
                onError={(syntheticEvent) => {
                  const { nativeEvent } = syntheticEvent
                  console.error('WebView error:', nativeEvent)
                  Alert.alert('Terminal Error', 'Failed to load terminal interface.')
                }}
                onHttpError={(syntheticEvent) => {
                  const { nativeEvent } = syntheticEvent
                  console.error('HTTP error:', nativeEvent)
                }}
                scalesPageToFit={false}
                javaScriptEnabled={true}
                domStorageEnabled={true}
                mixedContentMode="always"
                allowsFullscreenVideo={true}
                allowsInlineMediaPlayback={true}
                mediaPlaybackRequiresUserAction={false}
              />
            ) : (
              <View style={styles.loadingContainer}>
                <CardTitle>Loading terminal...</CardTitle>
              </View>
            )}
          </View>
          
          <View style={styles.controls}>
            <Button 
              title="Resize" 
              onPress={handleResize} 
              variant="outline"
              size="sm"
            />
            <Button 
              title="Cancel" 
              onPress={handleCancel} 
              variant="destructive"
              size="sm"
            />
          </View>
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
  terminalContainer: {
    flex: 1,
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderRadius: 8,
    overflow: 'hidden',
    marginBottom: 16,
  },
  terminal: {
    flex: 1,
    backgroundColor: '#1e1e1e',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  controls: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 8,
  },
})
