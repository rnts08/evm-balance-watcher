import React from 'react';
import { View, Text, FlatList } from 'react-native';
import { ArrowUpRight, ArrowDownLeft, Activity } from 'lucide-react-native';
import { Transaction } from '../types';

interface TransactionListProps {
    transactions: Transaction[];
    address: string;
}

export const TransactionList: React.FC<TransactionListProps> = ({ transactions, address }) => {
    const renderItem = ({ item }: { item: Transaction }) => {
        const isOut = item.from.toLowerCase() === address.toLowerCase();

        return (
            <View className="flex-row items-center py-4 border-b border-border">
                <View className={`p-2 rounded-full mr-4 ${isOut ? 'bg-error/10' : 'bg-success/10'}`}>
                    {isOut ? (
                        <ArrowUpRight size={18} color="#ef4444" />
                    ) : (
                        <ArrowDownLeft size={18} color="#22c55e" />
                    )}
                </View>
                <View className="flex-1">
                    <Text className="text-foreground font-medium text-sm">
                        {isOut ? 'Sent to ' : 'Received from '}
                        {isOut ? item.to.slice(0, 8) + '...' : item.from.slice(0, 8) + '...'}
                    </Text>
                    <Text className="text-muted-foreground text-xs mt-0.5">
                        {new Date(item.timestamp).toLocaleString()}
                    </Text>
                </View>
                <View className="items-end">
                    <Text className={`font-bold ${isOut ? 'text-foreground' : 'text-success'}`}>
                        {isOut ? '-' : '+'}{item.value}
                    </Text>
                    <Text className="text-muted-foreground text-[10px] mt-0.5 uppercase tracking-tighter">
                        Confirmed
                    </Text>
                </View>
            </View>
        );
    };

    return (
        <View className="flex-1">
            <View className="flex-row justify-between items-center mb-4">
                <Text className="text-foreground text-lg font-bold">Recent Transactions</Text>
                <Activity size={16} color="#3b82f6" />
            </View>

            {transactions.length === 0 ? (
                <View className="py-12 items-center opacity-50">
                    <Text className="text-muted-foreground italic">No transactions found</Text>
                </View>
            ) : (
                <FlatList
                    data={transactions}
                    renderItem={renderItem}
                    keyExtractor={(item) => item.hash}
                    scrollEnabled={false}
                />
            )}
        </View>
    );
};
