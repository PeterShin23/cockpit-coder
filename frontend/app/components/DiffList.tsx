import React from 'react'
import { View, Text, ScrollView, StyleSheet, TextStyle } from 'react-native'
import { Card, CardContent, CardHeader, CardTitle } from './ui/Card'
import { Button } from './ui/Button'

export interface Hunk {
  startOld: number
  lenOld: number
  startNew: number
  lenNew: number
  lines: string[]
}

export interface Patch {
  file: string
  patch: string
  hunks: Hunk[]
}

export interface DiffListProps {
  patches: Patch[]
  onSelectHunk: (file: string, hunkIndex: number) => void
  selectedHunks: Set<string> // format: "file:hunkIndex"
}

export function DiffList({ patches, onSelectHunk, selectedHunks }: DiffListProps) {
  const renderHunk = (hunk: Hunk, hunkIndex: number, file: string) => {
    const hunkKey = `${file}:${hunkIndex}`
    const isSelected = selectedHunks.has(hunkKey)

    return (
      <View key={hunkIndex} style={styles.hunk}>
        <View style={styles.hunkHeader}>
          <Text style={styles.hunkInfo}>
            @{hunk.startOld},{hunk.lenOld} +{hunk.startNew},{hunk.lenNew}
          </Text>
          <Button
            title={isSelected ? 'Deselect' : 'Select'}
            onPress={() => onSelectHunk(file, hunkIndex)}
            variant={isSelected ? 'destructive' : 'default'}
            size="sm"
          />
        </View>
        <View style={styles.hunkLines}>
          {hunk.lines.map((line, lineIndex) => {
            let lineStyle: TextStyle = styles.line
            if (line.startsWith('+')) {
              lineStyle = StyleSheet.flatten([styles.line, styles.lineAdded])
            } else if (line.startsWith('-')) {
              lineStyle = StyleSheet.flatten([styles.line, styles.lineRemoved])
            } else if (line.startsWith('\\')) {
              lineStyle = StyleSheet.flatten([styles.line, styles.lineContext])
            }
            return (
              <Text key={lineIndex} style={lineStyle}>
                {line}
              </Text>
            )
          })}
        </View>
      </View>
    )
  }

  return (
    <ScrollView style={styles.container}>
      {patches.map((patch, patchIndex) => (
        <Card key={patchIndex}>
          <CardHeader>
            <CardTitle style={styles.fileName}>{patch.file}</CardTitle>
          </CardHeader>
          <CardContent>
            <View style={styles.patch}>
              {patch.hunks.map((hunk, hunkIndex) =>
                renderHunk(hunk, hunkIndex, patch.file)
              )}
            </View>
          </CardContent>
        </Card>
      ))}
    </ScrollView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  fileName: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 8,
  },
  patch: {
    marginTop: 8,
  },
  hunk: {
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderRadius: 6,
    overflow: 'hidden',
  },
  hunkHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#F9FAFB',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  hunkInfo: {
    fontSize: 12,
    fontFamily: 'monospace',
    color: '#6B7280',
  },
  hunkLines: {
    padding: 8,
  },
  line: {
    fontSize: 13,
    fontFamily: 'monospace',
    lineHeight: 20,
    marginBottom: 2,
  },
  lineAdded: {
    color: '#059669',
    backgroundColor: '#D1FAE5',
  },
  lineRemoved: {
    color: '#DC2626',
    backgroundColor: '#FEE2E2',
  },
  lineContext: {
    color: '#6B7280',
  },
})
