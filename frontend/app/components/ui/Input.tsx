import React from 'react'
import { TextInput, StyleSheet } from 'react-native'
import { useThemeColor } from './useThemeColor'

export interface InputProps {
  value: string
  onChangeText: (text: string) => void
  placeholder?: string
  secureTextEntry?: boolean
  keyboardType?: any
  className?: string
  style?: any
}

export function Input({
  value,
  onChangeText,
  placeholder,
  secureTextEntry = false,
  keyboardType,
  className = '',
  style = {},
}: InputProps) {
  const backgroundColor = useThemeColor(
    { light: '#F9FAFB', dark: '#2C2C2E' },
    'input'
  )
  const textColor = useThemeColor(
    { light: '#111827', dark: '#F9FAFB' },
    'foreground'
  )
  const placeholderColor = useThemeColor(
    { light: '#9CA3AF', dark: '#6B7280' },
    'muted'
  )

  return (
    <TextInput
      style={[
        styles.input,
        {
          backgroundColor,
          color: textColor,
          borderColor: useThemeColor(
            { light: '#E5E7EB', dark: '#4B5563' },
            'border'
          ),
        },
        style,
        className,
      ]}
      value={value}
      onChangeText={onChangeText}
      placeholder={placeholder}
      placeholderTextColor={placeholderColor}
      secureTextEntry={secureTextEntry}
      keyboardType={keyboardType}
    />
  )
}

const styles = StyleSheet.create({
  input: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    fontSize: 16,
    minWidth: 200,
  },
})
