import React from 'react';
import { View, Text, ScrollView, RefreshControl, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { useStore } from '../../src/store/useStore';
import { BalanceCard } from '../../src/components/BalanceCard';
import { Activity, Bell, Wallet } from 'lucide-react-native';

export default function Dashboard() {
    const router = useRouter();
    const accounts = useStore((state) => state.accounts);
    const isConnected = useStore((state) => state.isConnected);
    const lastUpdate = useStore((state) => state.lastUpdate);

    const onRefresh = React.useCallback(() => {
        // Re-connect or trigger a status fetch
    }, []);

    return (
        <SafeAreaView className="flex-1 bg-background">
            <View className="px-6 py-4 flex-row justify-between items-center">
                <View>
                    <Text className="text-muted-foreground text-sm font-medium">Welcome back</Text>
                    <Text className="text-foreground text-2xl font-bold">Portfolio</Text>
                </View>
                <View className="flex-row gap-4">
                    <View className="bg-secondary p-3 rounded-full">
                        <Bell size={20} color="#fafafa" />
                    </View>
                    <View className="bg-secondary p-3 rounded-full">
                        <Activity size={20} color={isConnected ? "#22c55e" : "#ef4444"} />
                    </View>
                </View>
            </View>

            <ScrollView
                className="flex-1 px-6"
                showsVerticalScrollIndicator={false}
                refreshControl={
                    <RefreshControl refreshing={false} onRefresh={onRefresh} tintColor="#3b82f6" />
                }
            >
                <View className="py-2">
                    {accounts.length === 0 ? (
                        isConnected ? (
                            <View className="h-[60vh] items-center justify-center px-4">
                                <View className="bg-primary/10 p-6 rounded-full mb-6">
                                    <Wallet size={48} color="#3b82f6" />
                                </View>
                                <Text className="text-foreground text-2xl font-bold text-center mb-2">Welcome to EVM Bal</Text>
                                <Text className="text-muted-foreground text-center mb-8">
                                    Start monitoring your balances by adding your first wallet address.
                                </Text>
                                <TouchableOpacity
                                    onPress={() => router.push('/add-address')}
                                    className="bg-primary px-8 py-4 rounded-2xl shadow-lg shadow-primary/30"
                                >
                                    <Text className="text-white font-bold text-lg">Add Your First Address</Text>
                                </TouchableOpacity>
                            </View>
                        ) : (
                            <View className="h-64 items-center justify-center opacity-50">
                                <Text className="text-muted-foreground text-center">
                                    Connecting to backend...{"\n"}
                                    If this takes too long, check if the server is running.
                                </Text>
                            </View>
                        )
                    ) : (
                        accounts.map((acc, idx) => (
                            <BalanceCard key={acc.address + idx} account={acc} />
                        ))
                    )}
                </View>

                {lastUpdate && (
                    <Text className="text-muted-foreground text-[10px] text-center mb-8 uppercase tracking-widest">
                        Last updated: {lastUpdate.toLocaleTimeString()}
                    </Text>
                )}
            </ScrollView>
        </SafeAreaView>
    );
}
