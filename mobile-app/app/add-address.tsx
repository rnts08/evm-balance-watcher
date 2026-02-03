import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, KeyboardAvoidingView, Platform, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { X, Check, Wallet, Type } from 'lucide-react-native';

export default function AddAddressScreen() {
    const router = useRouter();
    const [address, setAddress] = useState('');
    const [name, setName] = useState('');

    const handleSave = () => {
        // In a real app we'd call apiClient.addAddress
        console.log('Saving address:', { address, name });
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
                            <Wallet size={16} color="#a3a3a3" className="mr-2" />
                            <Text className="text-muted-foreground text-xs uppercase font-bold tracking-wider">Wallet Address</Text>
                        </View>
                        <TextInput
                            className="bg-secondary text-foreground p-4 rounded-xl border border-border focus:border-primary"
                            placeholder="0x..."
                            placeholderTextColor="#525252"
                            value={address}
                            onChangeText={setAddress}
                            autoCapitalize="none"
                            autoCorrect={false}
                        />
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
                        disabled={!address}
                        className={`p-4 rounded-2xl items-center shadow-lg ${address ? 'bg-primary shadow-primary/30' : 'bg-secondary opacity-50'}`}
                    >
                        <Text className="text-white font-bold text-lg">Save Address</Text>
                    </TouchableOpacity>
                </ScrollView>
            </KeyboardAvoidingView>
        </SafeAreaView>
    );
}
