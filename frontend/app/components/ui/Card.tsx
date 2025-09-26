import React from 'react'
import { View, Text, StyleSheet } from 'react-native'
import { useThemeColor } from './useThemeColor'

export interface CardProps {
  children: React.ReactNode
  className?: string
}

export function Card({ children, className = '' }: CardProps) {
  const backgroundColor = useThemeColor(
    { light: '#FFFFFF', dark: '#1C1C1E' },
    'card'
  )
  const borderColor = useThemeColor(
    { light: '#E5E5E7', dark: '#38383A' },
    'border'
  )

  return (
    <View style={[styles.card, { backgroundColor, borderColor }]} className={className}>
      {children}
    </View>
  )
}

export interface CardHeaderProps {
  children: React.ReactNode
  className?: string
}

export function CardHeader({ children, className = '' }: CardHeaderProps) {
  return <View style={styles.header} className={className}>{children}</View>
}

export interface CardTitleProps {
  children: React.ReactNode
  className?: string
  style?: any
}

export function CardTitle({ children, className = '', style = {} }: CardTitleProps) {
  return <Text style={[styles.title, style]} className={className}>{children}</Text>
}

export interface CardContentProps {
  children: React.ReactNode
  className?: string
}

export function CardContent({ children, className = '' }: CardContentProps) {
  return <View style={styles.content} className={className}>{children}</View>
}

export interface CardFooterProps {
  children: React.ReactNode
  className?: string
}

export function CardFooter({ children, className = '' }: CardFooterProps) {
  return <View style={styles.footer} className={className}>{children}</View>
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 12,
    borderWidth: 1,
    padding: 16,
    marginVertical: 8,
    marginHorizontal: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  header: {
    marginBottom: 12,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 8,
  },
  content: {
    flex: 1,
  },
  footer: {
    marginTop: 12,
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
})
