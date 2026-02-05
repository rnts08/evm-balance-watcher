import { create } from 'zustand';
import { Account, ChainConfig, WatcherEvent, EventType } from '../types';

interface AppState {
    accounts: Account[];
    availableChains: ChainConfig[];
    prices: Record<string, number>;
    gasPrices: Record<string, string>;
    lastUpdate: Date | null;
    isConnected: boolean;

    // Actions
    setAccounts: (accounts: Account[]) => void;
    setAvailableChains: (chains: ChainConfig[]) => void;
    setPrices: (prices: Record<string, number>) => void;
    setIsConnected: (connected: boolean) => void;
    updateFromEvent: (event: WatcherEvent) => void;
}

const DEFAULT_CHAINS: ChainConfig[] = [
    {
        name: 'Ethereum',
        symbol: 'ETH',
        rpc_urls: [],
        explorer_url: 'https://etherscan.io',
        coingecko_id: 'ethereum',
        tokens: []
    },
    {
        name: 'Solana',
        symbol: 'SOL',
        rpc_urls: [],
        explorer_url: 'https://solscan.io',
        coingecko_id: 'solana',
        tokens: []
    },
    {
        name: 'Base',
        symbol: 'ETH',
        rpc_urls: [],
        explorer_url: 'https://basescan.org',
        coingecko_id: 'base',
        tokens: []
    },
    {
        name: 'Arbitrum',
        symbol: 'ETH',
        rpc_urls: [],
        explorer_url: 'https://arbiscan.io',
        coingecko_id: 'arbitrum-one',
        tokens: []
    }
];

export const useStore = create<AppState>((set) => ({
    accounts: [],
    availableChains: DEFAULT_CHAINS,
    prices: {},
    gasPrices: {},
    lastUpdate: null,
    isConnected: false,

    setAccounts: (accounts) => set({ accounts }),
    setAvailableChains: (availableChains) => set({ availableChains }),
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
