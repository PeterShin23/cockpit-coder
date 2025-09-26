import { Stack } from 'expo-router/stack'
import { StatusBar } from 'expo-status-bar'
import { GestureHandlerRootView } from 'react-native-gesture-handler'

export default function RootLayout() {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <StatusBar style="auto" />
      <Stack>
        <Stack.Screen name="index" options={{ headerShown: false }} />
        <Stack.Screen name="chat" options={{ headerShown: false }} />
        <Stack.Screen name="diffs" options={{ headerShown: false }} />
        <Stack.Screen name="commands" options={{ headerShown: false }} />
      </Stack>
    </GestureHandlerRootView>
  )
}
