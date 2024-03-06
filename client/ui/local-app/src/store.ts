import { writable } from 'svelte/store';
import type { AddressEntry, WalletMetadata } from '~/types';

// local API client
const apiClient = {
  // check if local API server is ready
  ready: async () => {
    try {
      const response = await fetch('http://localhost:9090/ready');
      return response.status === 200;
    } catch {
      return false;
    }
  },
  // get list of registered DIDs
  list: async () => {
    try {
      const response = await fetch('http://localhost:9090/list');
      const data = await response.json();
      return data;
    } catch {
      return [];
    }
  },
  // register a new DID
  register: async (name: string, recovery_key: string) => {
    try {
      const response = await fetch('http://localhost:9090/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name,
          recovery_key
        })
      });
      return response.status === 200;
    } catch {
      return false;
    }
  },
  // update DID's addresses
  update: async (
    name: string,
    did: string,
    addresses: AddressEntry[],
    passphrase: string
  ) => {
    try {
      const response = await fetch('http://localhost:9090/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name,
          did,
          passphrase,
          addresses
        })
      });
      return response.status === 200;
    } catch {
      return false;
    }
  }
};

// initialize state's store
const { subscribe, update } = writable({
  identifiers: [],
  wallet: {
    url: '',
    name: '',
    icons: [],
    addresses: [],
    connected: false,
    description: ''
  }
});

// application state handler
export const appState = {
  subscribe,
  reload: async function () {
    const list = await apiClient.list();
    list.sort((a, b) => {
      return Date.parse(b.last_sync) - Date.parse(a.last_sync);
    });
    update((st) => {
      st.identifiers = list;
      return st;
    });
  },
  setWallet: (md: WalletMetadata) => {
    update((st) => {
      st.wallet = md;
      return st;
    });
  },
  removeWallet: () => {
    update((st) => {
      st.wallet = {
        connected: false,
        addresses: [],
        name: '',
        description: '',
        url: '',
        icons: []
      };
      return st;
    });
  },
  createDID: async function (name: string, recovery_key: string): Promise<boolean> {
    const result = await apiClient.register(name, recovery_key);
    if (result) {
      this.reload();
    }
    return result;
  },
  updateDID: async function (
    name: string,
    did: string,
    addresses: AddressEntry[],
    passphrase: string
  ) {
    const result = await apiClient.update(name, did, addresses, passphrase);
    if (result) {
      this.reload();
    }
    return result;
  }
};
