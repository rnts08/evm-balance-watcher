import { useStore } from '../store/useStore';
import { WatcherEvent } from '../types';

const WS_URL = 'ws://localhost:8080/ws';

let socket: WebSocket | null = null;

export const connectWebSocket = () => {
    if (socket) return;

    socket = new WebSocket(WS_URL);

    socket.onopen = () => {
        console.log('WebSocket connected');
        useStore.getState().setIsConnected(true);
    };

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (data.type === 'initial') {
                useStore.getState().setAccounts(data.data.accounts);
                useStore.getState().setPrices(data.data.prices);
            } else {
                useStore.getState().updateFromEvent(data as WatcherEvent);
            }
        } catch (e) {
            console.error('Error parsing WebSocket message', e);
        }
    };

    socket.onclose = () => {
        console.log('WebSocket disconnected');
        useStore.getState().setIsConnected(false);
        socket = null;
        // Reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
    };

    socket.onerror = (error) => {
        console.error('WebSocket error', error);
    };
};

export const disconnectWebSocket = () => {
    if (socket) {
        socket.close();
        socket = null;
    }
};
