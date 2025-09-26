import React, { useState, useEffect } from 'react'
import { View, Text, StyleSheet, Alert, ScrollView } from 'react-native'
import { Button } from '../components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Input } from '../components/ui/Input'
import { apiClient } from '../lib/api'
import { getConnectionInfo } from '../lib/storage'

interface CommandsScreenProps {
  navigation?: any
}

interface Command {
  id: string
  cmd: string
  cwd: string
  timeoutMs: number
  status: 'pending' | 'running' | 'completed' | 'failed'
  output?: string
  createdAt: string
}

export function CommandsScreen({ navigation }: CommandsScreenProps) {
  const [commands, setCommands] = useState<Command[]>([])
  const [selectedCommand, setSelectedCommand] = useState<string>('')
  const [customCommand, setCustomCommand] = useState('')
  const [customCwd, setCustomCwd] = useState('')
  const [customTimeout, setCustomTimeout] = useState('60000')
  const [loading, setLoading] = useState(false)
  const [connectionInfo, setConnectionInfo] = useState<any>(null)

  // Allow-listed commands
  const allowListedCommands = [
    { cmd: 'npm test', cwd: '/path/to/repo', timeout: 300000 },
    { cmd: 'go test', cwd: '/path/to/repo', timeout: 120000 },
    { cmd: 'npm run build', cwd: '/path/to/repo', timeout: 300000 },
    { cmd: 'pytest', cwd: '/path/to/repo', timeout: 180000 },
    { cmd: 'git status', cwd: '/path/to/repo', timeout: 30000 },
    { cmd: 'ls -la', cwd: '/path/to/repo', timeout: 30000 },
  ]

  useEffect(() => {
    const loadConnectionInfo = async () => {
      try {
        const info = await getConnectionInfo()
        if (info) {
          setConnectionInfo(info)
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
  }, [navigation])

  const loadCommands = async () => {
    // This would typically fetch commands from the API
    // For now, we'll use mock data
    setCommands([
      {
        id: 'cmd_1',
        cmd: 'npm test',
        cwd: '/path/to/repo',
        timeoutMs: 300000,
        status: 'completed',
        output: 'Test results: 23 passing, 0 failing',
        createdAt: new Date().toISOString(),
      },
      {
        id: 'cmd_2',
        cmd: 'git status',
        cwd: '/path/to/repo',
        timeoutMs: 30000,
        status: 'running',
        createdAt: new Date().toISOString(),
      },
    ])
  }

  const handleRunCommand = async (cmd: string, cwd: string, timeoutMs: number) => {
    setLoading(true)
    try {
      const response = await apiClient.runCommand(cmd, cwd, timeoutMs)
      
      // Add to commands list
      const newCommand: Command = {
        id: `cmd_${Date.now()}`,
        cmd,
        cwd,
        timeoutMs,
        status: 'running',
        createdAt: new Date().toISOString(),
      }
      
      setCommands(prev => [newCommand, ...prev])
      
      // Simulate command completion
      setTimeout(() => {
        setCommands(prev => prev.map(c => 
          c.id === newCommand.id 
            ? { ...c, status: 'completed', output: 'Command completed successfully' }
            : c
        ))
      }, 2000)
      
    } catch (error) {
      console.error('Error running command:', error)
      Alert.alert('Error', 'Failed to run command.')
    } finally {
      setLoading(false)
    }
  }

  const handleRunCustomCommand = async () => {
    if (!customCommand.trim()) {
      Alert.alert('Invalid Command', 'Please enter a command to run.')
      return
    }

    const timeout = parseInt(customTimeout) || 60000
    handleRunCommand(customCommand.trim(), customCwd.trim() || '/path/to/repo', timeout)
    
    // Clear inputs
    setCustomCommand('')
    setCustomCwd('')
    setCustomTimeout('60000')
  }

  const getStatusColor = (status: Command['status']) => {
    switch (status) {
      case 'pending': return '#6B7280'
      case 'running': return '#059669'
      case 'completed': return '#059669'
      case 'failed': return '#DC2626'
      default: return '#6B7280'
    }
  }

  const getStatusText = (status: Command['status']) => {
    switch (status) {
      case 'pending': return 'Pending'
      case 'running': return 'Running'
      case 'completed': return 'Completed'
      case 'failed': return 'Failed'
      default: return 'Unknown'
    }
  }

  return (
    <View style={styles.container}>
      <Card>
        <CardHeader>
          <CardTitle>Commands</CardTitle>
        </CardHeader>
        <CardContent>
          {/* Allow-listed Commands */}
          <View style={styles.section}>
            <CardTitle style={styles.sectionTitle}>Quick Commands</CardTitle>
            <ScrollView style={styles.commandList}>
              {allowListedCommands.map((cmd, index) => (
                <View key={index} style={styles.commandItem}>
                  <View style={styles.commandInfo}>
                    <Text style={styles.commandText}>{cmd.cmd}</Text>
                    <Text style={styles.cwdText}>CWD: {cmd.cwd}</Text>
                  </View>
                  <Button
                    title="Run"
                    onPress={() => handleRunCommand(cmd.cmd, cmd.cwd, cmd.timeout)}
                    disabled={loading}
                    size="sm"
                  />
                </View>
              ))}
            </ScrollView>
          </View>

          {/* Custom Command */}
          <View style={styles.section}>
            <CardTitle style={styles.sectionTitle}>Custom Command</CardTitle>
            <View style={styles.customCommand}>
              <Input
                value={customCommand}
                onChangeText={setCustomCommand}
                placeholder="Enter command (e.g., npm start)"
                style={styles.input}
              />
              <Input
                value={customCwd}
                onChangeText={setCustomCwd}
                placeholder="Working directory (optional)"
                style={styles.input}
              />
              <Input
                value={customTimeout}
                onChangeText={setCustomTimeout}
                placeholder="Timeout (ms)"
                keyboardType="numeric"
                style={styles.input}
              />
              <Button
                title="Run Command"
                onPress={handleRunCustomCommand}
                disabled={loading || !customCommand.trim()}
                size="sm"
              />
            </View>
          </View>

          {/* Command History */}
          <View style={styles.section}>
            <CardTitle style={styles.sectionTitle}>Command History</CardTitle>
            <ScrollView style={styles.historyList}>
              {commands.map((cmd) => (
                <View key={cmd.id} style={styles.historyItem}>
                  <View style={styles.historyInfo}>
                    <Text style={styles.historyCommand}>{cmd.cmd}</Text>
                    <Text style={styles.historyCwd}>CWD: {cmd.cwd}</Text>
                    <View style={styles.historyMeta}>
                      <Text style={[styles.historyStatus, { color: getStatusColor(cmd.status) }]}>
                        {getStatusText(cmd.status)}
                      </Text>
                      <Text style={styles.historyTime}>
                        {new Date(cmd.createdAt).toLocaleTimeString()}
                      </Text>
                    </View>
                    {cmd.output && (
                      <Text style={styles.historyOutput}>{cmd.output}</Text>
                    )}
                  </View>
                </View>
              ))}
              
              {commands.length === 0 && (
                <View style={styles.emptyState}>
                  <Text style={styles.emptyText}>No commands run yet.</Text>
                </View>
              )}
            </ScrollView>
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
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 12,
    color: '#374151',
  },
  commandList: {
    maxHeight: 200,
  },
  commandItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
    backgroundColor: '#ffffff',
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  commandInfo: {
    flex: 1,
  },
  commandText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#111827',
    marginBottom: 4,
  },
  cwdText: {
    fontSize: 12,
    color: '#6B7280',
  },
  customCommand: {
    gap: 12,
  },
  input: {
    marginBottom: 8,
  },
  historyList: {
    flex: 1,
  },
  historyItem: {
    padding: 12,
    backgroundColor: '#ffffff',
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  historyInfo: {
    gap: 4,
  },
  historyCommand: {
    fontSize: 14,
    fontWeight: '500',
    color: '#111827',
  },
  historyCwd: {
    fontSize: 12,
    color: '#6B7280',
  },
  historyMeta: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 4,
  },
  historyStatus: {
    fontSize: 12,
    fontWeight: '500',
  },
  historyTime: {
    fontSize: 12,
    color: '#6B7280',
  },
  historyOutput: {
    fontSize: 12,
    color: '#374151',
    marginTop: 4,
    fontFamily: 'monospace',
    backgroundColor: '#F9FAFB',
    padding: 8,
    borderRadius: 4,
  },
  emptyState: {
    padding: 20,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    color: '#6B7280',
    textAlign: 'center',
  },
})
