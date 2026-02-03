import React from 'react';
import { View, Text, FlatList, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useStore } from '../../src/store/useStore';
import { User, Plus, Copy, Trash2 } from 'lucide-react-native';
import { useRouter } from 'expo-router';

export default function AddressesScreen() {
    const accounts = useStore((state) => state.accounts);
    const router = useRouter();

    const renderItem = ({ item }: { item: any }) => (
        <View className="bg-card p-4 rounded-2xl border border-border mb-3 flex-row items-center">
            <View className="bg-primary/10 p-3 rounded-xl mr-4">
                <User size={20} color="#3b82f6" />
            </View>
            <View className="flex-1">
                <Text className="text-foreground font-semibold text-base">{item.name || 'Unnamed Account'}</Text>
                <Text className="text-muted-foreground text-xs font-mono">{item.address.slice(0, 18)}...</Text>
            </View>
            <View className="flex-row gap-2">
                <TouchableOpacity className="bg-secondary p-2 rounded-lg">
                    <Copy size={16} color="#a3a3a3" />
                </TouchableOpacity>
                <TouchableOpacity className="bg-error/10 p-2 rounded-lg">
                    <Trash2 size={16} color="#ef4444" />
                </TouchableOpacity>
            </View>
        </View>
    );

    return (
        <SafeAreaView className="flex-1 bg-background">
            <View className="px-6 py-4 flex-row justify-between items-center">
                <View>
                    <Text className="text-foreground text-2xl font-bold">Addresses</Text>
                    <Text className="text-muted-foreground text-sm">Wallets being monitored</Text>
                </View>
                <TouchableOpacity
                    className="bg-primary p-3 rounded-full shadow-lg shadow-primary/30"
                    onPress={() => router.push('/add-address')}
                >
                    <Plus size={20} color="#ffffff" />
                </TouchableOpacity>
            </View>

            <FlatList
                data={accounts}
                renderItem={renderItem}
                keyExtractor={(item) => item.address}
                className="px-6 pt-2"
                ListEmptyComponent={
                    <View className="h-64 items-center justify-center opacity-50">
                        <User size={48} color="#525252" className="mb-4" />
                        <Text className="text-muted-foreground text-center">No addresses added yet</Text>
                    </View>
                }
            />
        </SafeAreaView>
    );
}
