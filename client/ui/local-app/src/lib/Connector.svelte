<script lang="ts">
  import WalletConnect from '@walletconnect/client';
  import QRCodeModal from 'algorand-walletconnect-qrcode-modal';
  import { createEventDispatcher, onMount } from 'svelte';
  import { appState } from '~/store';
  import { GetContext } from '~/context';
  import type { WalletMetadata } from '~/types';

  // component properties
  export let mainnet = false;

  // component event dispatcher
  // will be used to dispatch the events:
  //  - ready: wallet is connected and ready to use
  //  - session_end: wallet was disconnected
  //  - session_update: wallet details where updated
  const dispatch = createEventDispatcher();

  // get main application context
  const appCtx = GetContext();

  // connector instance
  let connector: WalletConnect;

  // setup the connector instance on component mount
  onMount(() => {
    // setup the connector instance
    connector = new WalletConnect({
      bridge: 'https://bridge.walletconnect.org',
      qrcodeModal: QRCodeModal,
      clientMeta: {
        name: 'AlgoID Connect (beta)',
        description: 'AlgoID wallet connector',
        url: 'http://github.com/algorandfoundation/did-algo',
        icons: [
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_128x128.png',
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_192x192.png',
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_512x512.png'
        ]
      }
    });

    // no need to continue if no session is established
    if (!connector.connected) {
      return;
    }

    // recover existing session
    let wallet: WalletMetadata = connector.peerMeta as WalletMetadata;
    wallet.connected = true;
    appState.setWallet(wallet);
    appCtx.showAlert('success', `Connected to: ${wallet.name}`);

    // start event processing
    dispatch('ready');
    handleConnectorEvents();
  });

  // setup the main wc connector instance
  async function startSession() {
    if (!connector.connected) {
      resetConnector();
      await connector.createSession({
        chainId: mainnet ? 416001 : 416002
      });
    }
    handleConnectorEvents();
  }

  // manually terminate a wc session, if previously established
  function endSession(): void {
    if (connector && connector.connected) {
      connector.killSession();
    }
  }

  // handle wc session events
  function handleConnectorEvents() {
    connector.on('connect', (error, payload) => {
      if (error) {
        appCtx.showAlert('error', `Error connecting to wallet: ${error.message}`);
        throw error;
      }
      let data = payload.params[0];
      let wallet = data.peerMeta;
      wallet.connected = true;
      wallet.addresses = data.accounts;
      appState.setWallet(wallet);
      appCtx.showAlert('success', `Connected to: ${wallet.name}`);
      dispatch('ready');
    });

    connector.on('disconnect', (error) => {
      if (error) {
        appCtx.showAlert('error', `Error disconnecting wallet: ${error.message}`);
        throw error;
      }
      appState.removeWallet();
      appCtx.showAlert('warning', `Disconnected from wallet`);
      dispatch('session_end');
    });

    connector.on('session_update', (error, payload) => {
      if (error) {
        appCtx.showAlert('error', `Wallet Connect session error: ${error.message}`);
        throw error;
      }
      let data = payload.params[0];
      let wallet = data.peerMeta;
      wallet.connected = true;
      wallet.addresses = data.accounts;
      appState.setWallet(wallet);
      dispatch('session_update');
    });
  }

  // reset the wc connector instance
  function resetConnector() {
    connector.off('connect');
    connector.off('disconnect');
    connector.off('session_update');
    connector = null;
    connector = new WalletConnect({
      bridge: 'https://bridge.walletconnect.org',
      qrcodeModal: QRCodeModal,
      clientMeta: {
        name: 'AlgoID Connect (beta)',
        description: 'AlgoID wallet connector',
        url: 'http://github.com/algorandfoundation/did-algo',
        icons: [
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_128x128.png',
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_192x192.png',
          'https://aid-tech.sfo2.digitaloceanspaces.com/public_assets/at_logo_512x512.png'
        ]
      }
    });
  }
</script>

<span class="absolute right-0 isolate inline-flex rounded-md shadow-sm">
  {#if !$appState.wallet.connected}
    <button
      on:click={startSession}
      type="button"
      class="inline-flex items-center gap-x-2 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500">
      <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none">
        <path
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M19.25 8.25V17.25C19.25 18.3546 18.3546 19.25 17.25 19.25H6.75C5.64543 19.25 4.75 18.3546 4.75 17.25V6.75" />
        <path
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M16.5 13C16.5 13.2761 16.2761 13.5 16 13.5C15.7239 13.5 15.5 13.2761 15.5 13C15.5 12.7239 15.7239 12.5 16 12.5C16.2761 12.5 16.5 12.7239 16.5 13Z" />
        <path
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M17.25 8.25H6.5C5.5335 8.25 4.75 7.4665 4.75 6.5C4.75 5.5335 5.5335 4.75 6.5 4.75H15.25C16.3546 4.75 17.25 5.64543 17.25 6.75V8.25ZM17.25 8.25H19.25" />
      </svg>
      Connect Wallet
    </button>
  {:else}
    <button
      on:click={endSession}
      type="button"
      class="inline-flex items-center gap-x-2 rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white hover:bg-red-500">
      {#if $appState.wallet.icons.length >= 1}
        <img class="h-6 w-6" alt="Wallet icon" src={$appState.wallet.icons[0]} />
      {/if}
      Disconnect: {$appState.wallet.name}
    </button>
  {/if}
</span>
