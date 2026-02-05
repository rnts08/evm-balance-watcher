import React, { useState, useEffect } from 'react';
import { View, Text, TextInput, TouchableOpacity, KeyboardAvoidingView, Platform, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { X, Check, Wallet, Type, Globe } from 'lucide-react-native';
import { useStore } from '../src/store/useStore';
import { ChainConfig } from '../src/types';

export default function AddAddressScreen() {
    const router = useRouter();
    const availableChains = useStore((state) => state.availableChains);
    const [address, setAddress] = useState('');
    const [name, setName] = useState('');
    const [selectedChain, setSelectedChain] = useState<ChainConfig | null>(null);
    const [isGuessing, setIsGuessing] = useState(false);

    // Guess chain based on address format
    useEffect(() => {
        if (address.length > 5) {
            setIsGuessing(true);
            if (address.startsWith('0x') && address.length === 42) {
                const eth = availableChains.find(c => c.name === 'Ethereum');
                if (eth) setSelectedChain(eth);
            } else if (/^[1-9A-HJ-NP-Za-km-z]{32,44}$/.test(address)) {
                const sol = availableChains.find(c => c.name === 'Solana');
                if (sol) setSelectedChain(sol);
            }
            setIsGuessing(false);
        }
    }, [address, availableChains]);

    const handleSave = () => {
        // In a real app we'd call apiClient.addAddress
        console.log('Saving address:', { address, name, chain: selectedChain?.name });
        router.back();
    };

    return (
        <SafeAreaView className="flex-1 bg-background">
            <View className="px-6 py-4 flex-row justify-between items-center border-b border-border">
                <TouchableOpacity onPress={() => router.back()} className="p-2 -ml-2">
                    <X size={24} color="#fafafa" />
                </TouchableOpacity>
                <Text className="text-foreground text-lg font-bold">Add Address</Text>
                <TouchableOpacity onPress={handleSave} className="p-2 -mr-2">
                    <Check size={24} color={address && selectedChain ? "#3b82f6" : "#525252"} />
                </TouchableOpacity>
            </View>

            <KeyboardAvoidingView
                behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
                className="flex-1"
            >
                <ScrollView className="px-6 py-8" showsVerticalScrollIndicator={false}>
                    <View className="mb-6">
                        <View className="flex-row items-center mb-2">
                            <Wallet size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Wallet Address</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="0x... or Solana address"
                            placeholderTextColor="#525252"
                            value={address}
                            onChangeText={setAddress}
                            autoCapitalize="none"
                            autoCorrect={false}
                        />
                    </View>

                    <View className="mb-6">
                        <View className="flex-row items-center mb-3">
                            <Globe size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Select Network</Text>
                        </View>
                        <View className="flex-row flex-wrap gap-2">
                            {availableChains.map((chain) => (
                                <TouchableOpacity
                                    key={chain.name}
                                    onPress={() => setSelectedChain(chain)}
                                    className={`px-4 py-2 rounded-full border ${selectedChain?.name === chain.name ? 'bg-primary/20 border-primary' : 'bg-secondary border-border'}`}
                                >
                                    <Text className={selectedChain?.name === chain.name ? 'text-primary font-bold' : 'text-muted-foreground'}>
                                        {chain.name}
                                    </Text>
                                </TouchableOpacity>
                            ))}
                        </View>
                        {selectedChain && (
                            <Text className="text-primary text-[10px] mt-2 italic font-medium uppercase tracking-widest text-center">
                                Suggested based on address format
                            </Text>
                        )}
                    </View>

                    <View className="mb-8">
                        <View className="flex-row items-center mb-2">
                            <Type size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Account Label</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="e.g. My Hot Wallet"
                            placeholderTextColor="#525252"
                            value={name}
                            onChangeText={setName}
                        />
                    </View>

                    <TouchableOpacity
                        onPress={handleSave}
                        disabled={!address || !selectedChain}
                        className={`p-4 rounded-2xl items-center shadow-lg ${address && selectedChain ? 'bg-primary shadow-primary/30' : 'bg-secondary opacity-50'}`}
                    >
                        <Text className="text-white font-bold text-lg">Save Address</Text>
                    </TouchableOpacity>
                </ScrollView>
            </KeyboardAvoidingView>
        </SafeAreaView>
    );
}
