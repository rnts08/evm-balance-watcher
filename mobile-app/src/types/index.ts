export interface ChainConfig {
    name: string;
    symbol: string;
    rpc_urls: string[];
    explorer_url: string;
    coingecko_id: string;
    tokens: TokenConfig[];
}

export interface TokenConfig {
    symbol: string;
    address: string;
    decimals: number;
    coingecko_id: string;
}

export interface Account {
    address: string;
    name: string;
    balances: Record<string, number>; // ChainName -> Balance
    token_balances: Record<string, Record<string, number>>; // ChainName -> Symbol -> Balance
    transactions: Transaction[];
}

export interface Transaction {
    hash: string;
    from: string;
    to: string;
    value: string;
    timestamp: string;
    status: string;
}

export interface WatcherEvent {
    type: string;
    data: any;
}

export const EventType = {
    PriceUpdated: "price_updated",
    ChainDataUpdated: "chain_data_updated",
    GasPriceUpdated: "gas_price_updated",
    TransactionsUpdated: "transactions_updated",
};
