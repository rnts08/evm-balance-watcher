import { create } from 'zustand';
import { Account, ChainConfig, WatcherEvent, EventType } from '../types';

interface AppState {
    accounts: Account[];
    chains: ChainConfig[];
    prices: Record<string, number>;
    gasPrices: Record<string, string>;
    lastUpdate: Date | null;
    isConnected: boolean;

    // Actions
    setAccounts: (accounts: Account[]) => void;
    setChains: (chains: ChainConfig[]) => void;
    setPrices: (prices: Record<string, number>) => void;
    setIsConnected: (connected: boolean) => void;
    updateFromEvent: (event: WatcherEvent) => void;
}

export const useStore = create<AppState>((set) => ({
    accounts: [],
    chains: [],
    prices: {},
    gasPrices: {},
    lastUpdate: null,
    isConnected: false,

    setAccounts: (accounts) => set({ accounts }),
    setChains: (chains) => set({ chains }),
    setPrices: (prices) => set({ prices }),
    setIsConnected: (isConnected) => set({ isConnected }),

    updateFromEvent: (event) => {
        set((state) => {
            switch (event.type) {
                case EventType.PriceUpdated:
                    return {
                        prices: { ...state.prices, [event.data.coin_id]: event.data.price },
                        lastUpdate: new Date(),
                    };
                case EventType.ChainDataUpdated:
                    // Update specific account balance if needed, or trigger a re-fetch
                    // For now, we'll rely on the server sending the full state periodically 
                    // or just update prices/gas in real-time.
                    return { lastUpdate: new Date() };
                case EventType.GasPriceUpdated:
                    return {
                        gasPrices: { ...state.gasPrices, [event.data.chain_name]: event.data.price },
                        lastUpdate: new Date(),
                    };
                default:
                    return state;
            }
        });
    },
}));
