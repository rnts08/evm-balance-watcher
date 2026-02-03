import axios from 'axios';
import { Account, ChainConfig } from '../types';

const BASE_URL = 'http://localhost:8080';

export const apiClient = {
    getStatus: async () => {
        const response = await axios.get(`${BASE_URL}/api/status`);
        return response.data;
    },

    // Placeholder for future config updates
    addAddress: async (address: string, name: string) => {
        // This would likely be a POST /api/config/address in a real app
        return axios.post(`${BASE_URL}/api/config/address`, { address, name });
    },
};
