import React from 'react';
import { View, Text, FlatList, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useStore } from '../../src/store/useStore';
import { Link2, Plus, Server, HardDrive } from 'lucide-react-native';
import { useRouter } from 'expo-router';

export default function ChainsScreen() {
    const chains = useStore((state) => state.chains);
    const router = useRouter();

    const renderItem = ({ item }: { item: any }) => (
        <View className="bg-card p-4 rounded-2xl border border-border mb-3 flex-row items-center">
            <View className="bg-accent/10 p-3 rounded-xl mr-4">
                <Server size={20} color="#a855f7" />
            </View>
            <View className="flex-1">
                <Text className="text-foreground font-semibold text-base">{item.name}</Text>
                <Text className="text-muted-foreground text-xs">{item.rpc_urls?.length || 0} RPC Endpoints</Text>
            </View>
            <View className="bg-secondary px-3 py-1 rounded-full">
                <Text className="text-muted-foreground text-[10px] font-bold uppercase">{item.symbol}</Text>
            </View>
        </View>
    );

    return (
        <SafeAreaView className="flex-1 bg-background">
            <View className="px-6 py-4 flex-row justify-between items-center">
                <View>
                    <Text className="text-foreground text-2xl font-bold">Chains</Text>
                    <Text className="text-muted-foreground text-sm">Manage network configurations</Text>
                </View>
                <TouchableOpacity
                    className="bg-primary p-3 rounded-full shadow-lg shadow-primary/30"
                    onPress={() => router.push('/add-chain')}
                >
                    <Plus size={20} color="#ffffff" />
                </TouchableOpacity>
            </View>

            <FlatList
                data={chains}
                renderItem={renderItem}
                keyExtractor={(item) => item.name}
                className="px-6 pt-2"
                ListEmptyComponent={
                    <View className="h-64 items-center justify-center opacity-50">
                        <Link2 size={48} color="#525252" className="mb-4" />
                        <Text className="text-muted-foreground text-center">No chains configured</Text>
                    </View>
                }
            />
        </SafeAreaView>
    );
}
