import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, KeyboardAvoidingView, Platform, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { X, Check, Link2, Type, Globe } from 'lucide-react-native';

export default function AddChainScreen() {
    const router = useRouter();
    const [name, setName] = useState('');
    const [symbol, setSymbol] = useState('');
    const [rpcUrl, setRpcUrl] = useState('');

    const handleSave = () => {
        // In a real app we'd call apiClient.addChain
        console.log('Saving chain:', { name, symbol, rpcUrl });
        router.back();
    };

    return (
        <SafeAreaView className="flex-1 bg-background">
            <View className="px-6 py-4 flex-row justify-between items-center border-b border-border">
                <TouchableOpacity onPress={() => router.back()} className="p-2 -ml-2">
                    <X size={24} color="#fafafa" />
                </TouchableOpacity>
                <Text className="text-foreground text-lg font-bold">Add Chain</Text>
                <TouchableOpacity onPress={handleSave} className="p-2 -mr-2">
                    <Check size={24} color="#3b82f6" />
                </TouchableOpacity>
            </View>

            <KeyboardAvoidingView
                behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
                className="flex-1"
            >
                <ScrollView className="px-6 py-8">
                    <View className="mb-6">
                        <View className="flex-row items-center mb-2">
                            <Link2 size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Network Name</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="e.g. Ethereum Mainnet"
                            placeholderTextColor="#525252"
                            value={name}
                            onChangeText={setName}
                        />
                    </View>

                    <View className="mb-6">
                        <View className="flex-row items-center mb-2">
                            <Type size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Symbol</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="e.g. ETH"
                            placeholderTextColor="#525252"
                            value={symbol}
                            onChangeText={setSymbol}
                            autoCapitalize="characters"
                        />
                    </View>

                    <View className="mb-8">
                        <View className="flex-row items-center mb-2">
                            <Globe size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">RPC URL</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="https://..."
                            placeholderTextColor="#525252"
                            value={rpcUrl}
                            onChangeText={setRpcUrl}
                            autoCapitalize="none"
                            autoCorrect={false}
                        />
                    </View>

                    <TouchableOpacity
                        onPress={handleSave}
                        disabled={!name || !symbol || !rpcUrl}
                        className={`p-4 rounded-2xl items-center shadow-lg ${name && symbol && rpcUrl ? 'bg-primary shadow-primary/30' : 'bg-secondary opacity-50'}`}
                    >
                        <Text className="text-white font-bold text-lg">Save Chain</Text>
                    </TouchableOpacity>
                </ScrollView>
            </KeyboardAvoidingView>
        </SafeAreaView>
    );
}
