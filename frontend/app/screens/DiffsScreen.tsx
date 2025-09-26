import React, { useState, useEffect } from 'react'
import { View, Text, StyleSheet, Alert, ScrollView } from 'react-native'
import { Button } from '../components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { DiffList } from '../components/DiffList'
import { apiClient } from '../lib/api'
import { getConnectionInfo } from '../lib/storage'

interface DiffsScreenProps {
  navigation?: any
}

export function DiffsScreen({ navigation }: DiffsScreenProps) {
  const [taskId, setTaskId] = useState<string>('')
  const [patches, setPatches] = useState<any[]>([])
  const [selectedHunks, setSelectedHunks] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)
  const [connectionInfo, setConnectionInfo] = useState<any>(null)

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

  const loadPatches = async () => {
    if (!taskId) {
      Alert.alert('No Task', 'Please select a task first.')
      return
    }

    setLoading(true)
    try {
      const response = await apiClient.getTaskPatches(taskId)
      setPatches(response.patches || [])
      setSelectedHunks(new Set())
    } catch (error) {
      console.error('Error loading patches:', error)
      Alert.alert('Error', 'Failed to load code differences.')
    } finally {
      setLoading(false)
    }
  }

  const handleSelectHunk = (file: string, hunkIndex: number) => {
    const hunkKey = `${file}:${hunkIndex}`
    const newSelected = new Set(selectedHunks)
    
    if (newSelected.has(hunkKey)) {
      newSelected.delete(hunkKey)
    } else {
      newSelected.add(hunkKey)
    }
    
    setSelectedHunks(newSelected)
  }

  const handleApplyPatches = async () => {
    if (selectedHunks.size === 0) {
      Alert.alert('No Selection', 'Please select at least one hunk to apply.')
      return
    }

    if (!taskId) {
      Alert.alert('No Task', 'Please select a task first.')
      return
    }

    setLoading(true)
    try {
      const selections: any[] = []
      selectedHunks.forEach(hunkKey => {
        const [file, hunkIndex] = hunkKey.split(':')
        selections.push({
          file,
          hunks: [parseInt(hunkIndex)]
        })
      })

      await apiClient.applyTaskPatches(taskId, selections, 'Apply selected hunks')
      Alert.alert('Success', 'Patches applied successfully!')
      setSelectedHunks(new Set())
      await loadPatches() // Refresh patches
    } catch (error) {
      console.error('Error applying patches:', error)
      Alert.alert('Error', 'Failed to apply patches.')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateTask = async () => {
    // This would typically open a dialog to create a new task
    // For now, we'll use a placeholder task ID
    const newTaskId = `task_${Date.now()}`
    setTaskId(newTaskId)
    
    try {
      // Create a sample task
      await apiClient.createTask(
        'Sample task instruction',
        'main',
        { files: [], hints: '' },
        'cline'
      )
      await loadPatches()
    } catch (error) {
      console.error('Error creating task:', error)
      Alert.alert('Error', 'Failed to create task.')
    }
  }

  return (
    <View style={styles.container}>
      <Card>
        <CardHeader>
          <CardTitle>Code Differences</CardTitle>
        </CardHeader>
        <CardContent>
          <View style={styles.taskSection}>
            <View style={styles.taskControls}>
              <Button 
                title={taskId ? 'Change Task' : 'Create Task'} 
                onPress={handleCreateTask}
                variant="outline"
                size="sm"
              />
              {taskId && (
                <Button 
                  title="Load Patches" 
                  onPress={loadPatches}
                  disabled={loading}
                  size="sm"
                />
              )}
            </View>
            
            {taskId && (
              <View style={styles.taskInfo}>
                <Text style={styles.taskIdText}>Task ID: {taskId}</Text>
              </View>
            )}
          </View>

          {loading && (
            <View style={styles.loadingContainer}>
              <Text>Loading...</Text>
            </View>
          )}

          {!loading && patches.length > 0 && (
            <View style={styles.diffSection}>
              <View style={styles.diffControls}>
                <Button 
                  title="Apply Selected" 
                  onPress={handleApplyPatches}
                  disabled={selectedHunks.size === 0}
                  size="sm"
                />
                <Button 
                  title="Clear Selection" 
                  onPress={() => setSelectedHunks(new Set())}
                  variant="outline"
                  size="sm"
                />
              </View>
              
              <ScrollView style={styles.diffList}>
                <DiffList
                  patches={patches}
                  onSelectHunk={handleSelectHunk}
                  selectedHunks={selectedHunks}
                />
              </ScrollView>
            </View>
          )}

          {!loading && patches.length === 0 && taskId && (
            <View style={styles.emptyState}>
              <Text style={styles.emptyText}>No patches available for this task.</Text>
            </View>
          )}

          {!taskId && (
            <View style={styles.emptyState}>
              <Text style={styles.emptyText}>No task selected. Create a task to see differences.</Text>
            </View>
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
  taskSection: {
    marginBottom: 16,
  },
  taskControls: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 8,
  },
  taskInfo: {
    padding: 8,
    backgroundColor: '#f9fafb',
    borderRadius: 4,
  },
  taskIdText: {
    fontSize: 12,
    color: '#6b7280',
    fontFamily: 'monospace',
  },
  loadingContainer: {
    padding: 20,
    alignItems: 'center',
  },
  diffSection: {
    flex: 1,
  },
  diffControls: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 16,
  },
  diffList: {
    flex: 1,
  },
  emptyState: {
    padding: 20,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    color: '#6b7280',
    textAlign: 'center',
  },
})
