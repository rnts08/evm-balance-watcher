import { useEffect } from 'react';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { connectWebSocket, disconnectWebSocket } from '../src/api/socket';
import { useStore } from '../src/store/useStore';
import '../global.css';

export default function RootLayout() {
    const isConnected = useStore((state) => state.isConnected);

    useEffect(() => {
        connectWebSocket();
        return () => disconnectWebSocket();
    }, []);

    return (
        <>
            <Stack
                screenOptions={{
                    headerShown: false,
                    contentStyle: { backgroundColor: '#09090b' },
                }}
            >
                <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
                <Stack.Screen
                    name="add-address"
                    options={{
                        presentation: 'modal',
                        animation: 'slide_from_bottom'
                    }}
                />
                <Stack.Screen
                    name="add-chain"
                    options={{
                        presentation: 'modal',
                        animation: 'slide_from_bottom'
                    }}
                />
            </Stack>
            <StatusBar style="light" />
        </>
    );
}
