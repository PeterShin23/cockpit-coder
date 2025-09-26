import { useColorScheme } from 'react-native'
import { useTheme } from '@react-navigation/native'

type ThemeProps = {
  light?: string
  dark?: string
}

export function useThemeColor(props: ThemeProps, colorName: string) {
  const theme = useColorScheme() ?? 'light'
  const { colors } = useTheme()
  
  const colorFromProps = props[theme as 'light' | 'dark']

  if (colorFromProps) {
    return colorFromProps
  } else {
    return colors[colorName as keyof typeof colors] || colorName
  }
}
