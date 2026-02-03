import React from 'react';
import { View, Text } from 'react-native';
import { Wallet, TrendingUp, ChevronRight } from 'lucide-react-native';
import { Account } from '../types';
import { useStore } from '../store/useStore';

interface BalanceCardProps {
    account: Account;
    onPress?: () => void;
}

export const BalanceCard: React.FC<BalanceCardProps> = ({ account, onPress }) => {
    const prices = useStore((state) => state.prices);

    // Calculate total USD value for this account
    const calculateTotalValue = () => {
        let total = 0;
        // Native balances
        Object.entries(account.balances).forEach(([chainName, balance]) => {
            // Find chain config to get coingecko_id (simplified for now)
            // In a real app we'd map chain name to CG ID
            const price = prices['ethereum'] || 0; // Defaulting to eth price for demo
            total += balance * price;
        });

        // Token balances
        Object.values(account.token_balances).forEach((tokens) => {
            Object.entries(tokens).forEach(([symbol, balance]) => {
                const price = prices[symbol.toLowerCase()] || 0;
                total += balance * price;
            });
        });

        return total;
    };

    const totalValue = calculateTotalValue();

    return (
        <View className="bg-card p-5 rounded-3xl border border-border mb-4 shadow-sm">
            <View className="flex-row justify-between items-center mb-4">
                <View className="flex-row items-center">
                    <View className="bg-primary/10 p-2 rounded-xl mr-3">
                        <Wallet size={20} color="#3b82f6" />
                    </View>
                    <View>
                        <Text className="text-foreground font-semibold text-lg">{account.name || 'Account'}</Text>
                        <Text className="text-muted-foreground text-xs font-mono">
                            {account.address.slice(0, 6)}...{account.address.slice(-4)}
                        </Text>
                    </View>
                </View>
                <ChevronRight size={20} color="#525252" />
            </View>

            <View className="mb-2">
                <Text className="text-muted-foreground text-xs uppercase tracking-wider mb-1">Total Value</Text>
                <Text className="text-foreground text-3xl font-bold">
                    ${totalValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </Text>
            </View>

            <View className="flex-row items-center">
                <TrendingUp size={14} color="#22c55e" />
                <Text className="text-success text-xs font-medium ml-1">+2.45% (24h)</Text>
            </View>
        </View>
    );
};
