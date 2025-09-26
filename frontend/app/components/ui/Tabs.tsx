import React from 'react'
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native'
import { useThemeColor } from './useThemeColor'

export interface TabsProps {
  children: React.ReactNode
  className?: string
}

export function Tabs({ children, className = '' }: TabsProps) {
  return <View style={styles.tabs} className={className}>{children}</View>
}

export interface TabsListProps {
  children: React.ReactNode
  className?: string
}

export function TabsList({ children, className = '' }: TabsListProps) {
  return <View style={styles.list} className={className}>{children}</View>
}

export interface TabsTriggerProps {
  children: React.ReactNode
  value: string
  isActive?: boolean
  onPress: () => void
  className?: string
}

export function TabsTrigger({
  children,
  value,
  isActive = false,
  onPress,
  className = '',
}: TabsTriggerProps) {
  const backgroundColor = useThemeColor(
    { light: '#F9FAFB', dark: '#2C2C2E' },
    'input'
  )
  const activeColor = useThemeColor(
    { light: '#007AFF', dark: '#0A84FF' },
    'primary'
  )
  const textColor = useThemeColor(
    { light: '#111827', dark: '#F9FAFB' },
    'foreground'
  )

  return (
    <TouchableOpacity
      style={[
        styles.trigger,
        {
          backgroundColor: isActive ? activeColor : backgroundColor,
          borderColor: useThemeColor(
            { light: '#E5E7EB', dark: '#4B5563' },
            'border'
          ),
        },
      ]}
      onPress={onPress}
      className={className}
    >
      <Text
        style={[
          styles.triggerText,
          { color: isActive ? '#FFFFFF' : textColor },
        ]}
      >
        {children}
      </Text>
    </TouchableOpacity>
  )
}

export interface TabsContentProps {
  children: React.ReactNode
  value: string
  className?: string
}

export function TabsContent({ children, className = '' }: TabsContentProps) {
  return <View style={styles.content} className={className}>{children}</View>
}

const styles = StyleSheet.create({
  tabs: {
    flexDirection: 'column',
  },
  list: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  trigger: {
    flex: 1,
    paddingVertical: 12,
    paddingHorizontal: 16,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 8,
    marginHorizontal: 4,
    borderWidth: 1,
    borderBottomWidth: 0,
  },
  triggerText: {
    fontSize: 14,
    fontWeight: '600',
  },
  content: {
    flex: 1,
    paddingVertical: 16,
  },
})
