import React, { useState, useEffect, useRef } from 'react'
import { View, Text, StyleSheet, Alert, ScrollView, TextInput } from 'react-native'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { getConnectionInfo } from '../lib/storage'
import { WebSocketProvider, useWebSocket } from '../components/WebSocketProvider'

interface Message {
  id: string
  type: 'user' | 'agent' | 'system' | 'confirmation'
  content: string
  timestamp: Date
  status?: 'pending' | 'sent' | 'received' | 'confirmed' | 'cancelled'
}

interface ChatScreenProps {
  navigation?: any
}

function ChatScreenContent({ navigation }: ChatScreenProps) {
  const { sendMessage, isConnected, messages: wsMessages } = useWebSocket()
  const [localMessages, setLocalMessages] = useState<Message[]>([])
  const [inputText, setInputText] = useState('')
  const scrollViewRef = useRef<ScrollView>(null)

  // Combine local messages with WebSocket messages
  const allMessages = [...localMessages, ...wsMessages.map((msg: any, index: number) => ({
    id: `ws-${index}`,
    type: msg.type === 'user' ? 'user' : 'agent',
    content: msg.content || JSON.stringify(msg),
    timestamp: new Date(),
    status: msg.status
  }))]

  useEffect(() => {
    const loadConnectionInfo = async () => {
      try {
        const info = await getConnectionInfo()
        if (!info) {
          Alert.alert('No Connection', 'Please scan a QR code to connect first.')
          navigation?.navigate('index')
        }
      } catch (error) {
        console.error('Error loading connection info:', error)
        Alert.alert('Connection Error', 'Failed to load connection information.')
      }
    }

    loadConnectionInfo()
  }, [navigation])

  useEffect(() => {
    // Add welcome message when connected
    if (isConnected && localMessages.length === 0) {
      addMessage({
        id: 'welcome',
        type: 'system',
        content: 'Connected to coding agent. Start a conversation or wait for agent updates.',
        timestamp: new Date()
      })
    }
  }, [isConnected])

  const addMessage = (message: Message) => {
    setLocalMessages(prev => [...prev, message])
    // Scroll to bottom
    setTimeout(() => {
      scrollViewRef.current?.scrollToEnd({ animated: true })
    }, 100)
  }

  const handleSendMessage = () => {
    if (!inputText.trim() || !isConnected) return

    const message: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputText,
      timestamp: new Date(),
      status: 'sent'
    }

    addMessage(message)
    setInputText('')

    // Send message to backend
    sendMessage({
      type: 'user_message',
      content: inputText,
      timestamp: message.timestamp.toISOString()
    })
  }

  const handleConfirmAction = (messageId: string) => {
    // Find the confirmation message and update its status
    setLocalMessages(prev => prev.map(msg => 
      msg.id === messageId 
        ? { ...msg, status: 'confirmed' }
        : msg
    ))

    // Send confirmation to backend
    sendMessage({
      type: 'confirmation',
      messageId: messageId,
      confirmed: true
    })
  }

  const handleCancelAction = (messageId: string) => {
    // Find the confirmation message and update its status
    setLocalMessages(prev => prev.map(msg => 
      msg.id === messageId 
        ? { ...msg, status: 'cancelled' }
        : msg
    ))

    // Send cancellation to backend
    sendMessage({
      type: 'confirmation',
      messageId: messageId,
      confirmed: false
    })
  }

  const handleDisconnect = () => {
    navigation?.navigate('index')
  }

  const renderMessage = (message: Message) => {
    const isUser = message.type === 'user'
    const isConfirmation = message.type === 'confirmation'
    
    return (
      <View 
        key={message.id}
        style={[
          styles.messageContainer,
          isUser ? styles.userMessage : styles.agentMessage,
          isConfirmation && styles.confirmationMessage
        ]}
      >
        <View style={[
          styles.messageBubble,
          isUser ? styles.userBubble : styles.agentBubble,
          isConfirmation && styles.confirmationBubble
        ]}>
          <Text style={[
            styles.messageText,
            isUser ? styles.userText : styles.agentText,
            isConfirmation && styles.confirmationText
          ]}>
            {message.content}
          </Text>
          
          {message.type === 'confirmation' && (
            <View style={styles.confirmationButtons}>
              <Button
                title="Confirm"
                onPress={() => handleConfirmAction(message.id)}
                size="sm"
                variant="default"
              />
              <Button
                title="Cancel"
                onPress={() => handleCancelAction(message.id)}
                size="sm"
                variant="destructive"
              />
            </View>
          )}
          
          <Text style={styles.timestamp}>
            {message.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </Text>
        </View>
      </View>
    )
  }

  return (
    <View style={styles.container}>
      <Card>
        <CardHeader>
          <View style={styles.header}>
            <CardTitle>Chat with Coding Agent</CardTitle>
            <View style={styles.connectionStatus}>
              <View style={[
                styles.statusIndicator,
                isConnected ? styles.connected : styles.disconnected
              ]} />
              <Text style={styles.statusText}>
                {isConnected ? 'Connected' : 'Disconnected'}
              </Text>
            </View>
          </View>
        </CardHeader>
        <CardContent>
          <ScrollView 
            ref={scrollViewRef}
            style={styles.messagesContainer}
            contentContainerStyle={styles.messagesContent}
          >
            {allMessages.map(renderMessage)}
          </ScrollView>
          
          <View style={styles.inputContainer}>
            <TextInput
              style={styles.input}
              value={inputText}
              onChangeText={setInputText}
              placeholder="Type your message..."
              multiline
              numberOfLines={3}
              maxLength={1000}
            />
            <Button
              title="Send"
              onPress={handleSendMessage}
              disabled={!inputText.trim() || !isConnected}
              size="sm"
            />
          </View>
          
          <View style={styles.actionButtons}>
            <Button
              title="View Diffs"
              onPress={() => navigation?.navigate('diffs')}
              variant="outline"
              size="sm"
            />
            <Button
              title="Run Commands"
              onPress={() => navigation?.navigate('commands')}
              variant="outline"
              size="sm"
            />
            <Button
              title="Disconnect"
              onPress={handleDisconnect}
              variant="destructive"
              size="sm"
            />
          </View>
        </CardContent>
      </Card>
    </View>
  )
}

export function ChatScreen({ navigation }: ChatScreenProps) {
  const [connectionInfo, setConnectionInfo] = useState<any>(null)

  useEffect(() => {
    const loadConnectionInfo = async () => {
      try {
        const info = await getConnectionInfo()
        setConnectionInfo(info)
      } catch (error) {
        console.error('Error loading connection info:', error)
      }
    }

    loadConnectionInfo()
  }, [])

  if (!connectionInfo?.ws?.events) {
    return (
      <View style={styles.container}>
        <Card>
          <CardContent>
            <Text>Loading connection...</Text>
          </CardContent>
        </Card>
      </View>
    )
  }

  return (
    <WebSocketProvider endpoint={connectionInfo.ws.events}>
      <ChatScreenContent navigation={navigation} />
    </WebSocketProvider>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  connectionStatus: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  statusIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  connected: {
    backgroundColor: '#10B981',
  },
  disconnected: {
    backgroundColor: '#EF4444',
  },
  statusText: {
    fontSize: 12,
    color: '#6B7280',
  },
  messagesContainer: {
    flex: 1,
    marginBottom: 16,
  },
  messagesContent: {
    paddingVertical: 8,
  },
  messageContainer: {
    marginVertical: 4,
    marginHorizontal: 8,
  },
  messageBubble: {
    padding: 12,
    borderRadius: 12,
    maxWidth: '80%',
  },
  userMessage: {
    alignItems: 'flex-end',
  },
  agentMessage: {
    alignItems: 'flex-start',
  },
  confirmationMessage: {
    alignItems: 'center',
    width: '100%',
  },
  userBubble: {
    backgroundColor: '#3B82F6',
    borderBottomRightRadius: 4,
  },
  agentBubble: {
    backgroundColor: '#E5E7EB',
    borderBottomLeftRadius: 4,
  },
  confirmationBubble: {
    backgroundColor: '#FEF3C7',
    alignItems: 'center',
    width: '100%',
    maxWidth: '100%',
  },
  messageText: {
    fontSize: 14,
    lineHeight: 20,
  },
  userText: {
    color: '#FFFFFF',
  },
  agentText: {
    color: '#1F2937',
  },
  confirmationText: {
    color: '#92400E',
    fontWeight: '600',
    marginBottom: 8,
  },
  timestamp: {
    fontSize: 10,
    color: '#9CA3AF',
    marginTop: 4,
    alignSelf: 'flex-end',
  },
  confirmationButtons: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 8,
  },
  inputContainer: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 16,
  },
  input: {
    flex: 1,
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderRadius: 8,
    padding: 12,
    backgroundColor: '#FFFFFF',
    fontSize: 14,
    maxHeight: 100,
  },
  actionButtons: {
    flexDirection: 'row',
    gap: 8,
    justifyContent: 'space-between',
  },
})
