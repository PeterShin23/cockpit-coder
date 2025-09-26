import React from 'react'
import { TouchableOpacity, Text, StyleSheet } from 'react-native'
import { useThemeColor } from './useThemeColor'

export interface ButtonProps {
  onPress: () => void
  title: string
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  disabled?: boolean
  className?: string
}

export function Button({
  onPress,
  title,
  variant = 'default',
  size = 'default',
  disabled = false,
  className = '',
}: ButtonProps) {
  const backgroundColor = useThemeColor(
    { light: '#007AFF', dark: '#0A84FF' },
    'primary'
  )
  const textColor = useThemeColor(
    { light: '#FFFFFF', dark: '#FFFFFF' },
    'primary'
  )

  const getButtonStyle = () => {
    const baseStyle = [
      styles.button,
      disabled && styles.disabled,
      variant === 'outline' && styles.outline,
      variant === 'destructive' && styles.destructive,
      size === 'sm' && styles.small,
      size === 'lg' && styles.large,
    ]

    if (variant === 'default' || variant === 'secondary') {
      return [...baseStyle, { backgroundColor: disabled ? '#999' : backgroundColor }]
    }

    return baseStyle
  }

  const getTextStyle = () => {
    const baseStyle = [styles.text]
    
    if (variant === 'outline' || variant === 'ghost') {
      return [...baseStyle, { color: backgroundColor }]
    }
    
    return [...baseStyle, { color: textColor }]
  }

  return (
    <TouchableOpacity
      style={getButtonStyle()}
      onPress={onPress}
      disabled={disabled}
      className={className}
    >
      <Text style={getTextStyle()}>{title}</Text>
    </TouchableOpacity>
  )
}

const styles = StyleSheet.create({
  button: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minWidth: 80,
  },
  text: {
    fontSize: 16,
    fontWeight: '600',
  },
  disabled: {
    opacity: 0.5,
  },
  outline: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#007AFF',
  },
  destructive: {
    backgroundColor: '#FF3B30',
  },
  small: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    minWidth: 60,
  },
  large: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    minWidth: 120,
  },
})
