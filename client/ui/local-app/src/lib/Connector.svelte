<script lang="ts">
  import {SignalClient} from '@algorandfoundation/liquid-client'
  import { createEventDispatcher, onMount } from 'svelte';
  import { appState } from '~/store';
  import { GetContext } from '~/context';
  import WalletModal from "~/lib/WalletModal.svelte";
  import type {ModalOptions} from "~/types";

  const dispatch = createEventDispatcher();

  // get main application context
  const appCtx = GetContext();

  // connector instance
  let client: SignalClient;

  // setup the SignalClient instance on component mount
  onMount(() => {
    console.log('mounting')
    client = new SignalClient("https://liquid-auth.onrender.com")
    handleConnectorEvents();
  });

  // peer with a wallet
  async function startSession() {
    const requestId = SignalClient.generateRequestId()
    if(!client.requestId){
      client.peer(requestId, "offer").then((dc)=>{
        dc.send('What up homie')
        appState.setDataChannel(dc)
    })
    }
    appCtx.showModal({
      asPanel: false,
      title: 'Connect Wallet',
      subtitle: 'Scan the QR code with your wallet app to connect',
      content: WalletModal,
      props: {
        src: await client.qrCode(),
        hidden: false
      }
    } as ModalOptions);
  }

  // manually terminate a session, if previously established
  function endSession(): void {
    client.close()
    appState.removeWallet();
    appCtx.showAlert('error', `Disconnected from wallet`);
  }

  // handle events
  function handleConnectorEvents() {
    client.on('link-message',(msg)=>{
      appState.setWallet({
        connected: true,
        addresses: [msg.wallet],
        name: `${msg.wallet.substring(0,4)}...${msg.wallet.substring(msg.wallet.length - 4, msg.wallet.length)}`,
        description: "WebRTC Wallet",
        url: "https://liquid-auth.onrender.com",
        icons: []
      });
      appCtx.closeModal()
      appCtx.showAlert('success', `Connected to: ${msg.wallet}`);
      dispatch('ready');
    })
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
