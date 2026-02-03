import React from 'react';
import { View, Text, ScrollView, RefreshControl } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useStore } from '../../src/store/useStore';
import { BalanceCard } from '../../src/components/BalanceCard';
import { Activity, Bell } from 'lucide-react-native';

export default function Dashboard() {
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
                        <View className="h-64 items-center justify-center opacity-50">
                            <Text className="text-muted-foreground text-center">
                                Connecting to backend...{"\n"}
                                If this takes too long, check if the server is running.
                            </Text>
                        </View>
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
