<script lang="ts">
  import Switch from '~/lib/Switch.svelte';
  import { GetContext } from '~/context';
  import { appState } from '~/store';
  import type { AddressEntry, IdentifierEntry } from '~/types';

  // component properties
  export let identifier: IdentifierEntry;

  // main application context
  const appCtx = GetContext();

  // available addresses
  // the reactivity block will be executed whenever `appState.wallet`
  // changes, automatically updating the list of available addresses.
  let addresses: AddressEntry[];
  $: {
    // clear list
    addresses = [];

    // add addresses currently associated with the identifier
    identifier.addresses.forEach((entry) => {
      addresses = [
        ...addresses,
        {
          address: entry.address,
          network: entry.network,
          enabled: true
        }
      ];
    });

    // add addresses available in the connected wallet
    $appState.wallet.addresses.forEach((address) => {
      if (!addresses.find((entry) => entry.address == address)) {
        addresses = [
          ...addresses,
          {
            address: address,
            network: 'mainnet',
            enabled: false
          }
        ];
      }
    });
  }

  // process data submission
  async function submit() {
    let result = await appState.updateDID(identifier.name, identifier.did, addresses);
    appCtx.closeModal();
    if (result) {
      appCtx.showAlert('success', 'Identifier updated successfully.');
    } else {
      appCtx.showAlert('error', 'Failed to updated identifier.');
    }
  }

  // format a crypto address to a common textual representation.
  function formatAddress(address: string): string {
    return address.slice(0, 16) + '...' + address.slice(-16);
  }

  // toggle the enabled state of an address entry
  function toggleAddress(entry: AddressEntry) {
    entry.enabled = !entry.enabled;
  }
</script>

<div class="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
  <!-- desc -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5 first:sm:pt-0">
    <p class="sm:col-span-3">
      Adjust the <strong>ALGO</strong> addresses associated with this identifier. You can add
      as many as you want. The DID document associated will be automatically synced with the
      network.
    </p>
  </div>
  <!-- address list -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    {#if addresses.length == 0}
      <p class="col-span-3 text-center text-gray-500">
        There are no available addresses, connect your wallet and try again!
      </p>
    {:else}
      <div class="col-span-3 flex flex-col items-center space-y-3">
        {#each addresses as entry}
          <div class="flex items-center space-x-2">
            <Switch
              enabled={entry.enabled}
              on:toggle={() => {
                toggleAddress(entry);
              }} />
            <span
              class="inline-flex rounded-full bg-indigo-100 px-4 py-2 text-base font-semibold leading-5 text-indigo-800 transition-colors"
              >{formatAddress(entry.address)}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
  <!-- action buttons -->
  <div class="flex-shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
    <div class="flex justify-end space-x-3">
      <button
        on:click={() => {
          appCtx.closeModal();
        }}
        type="button"
        class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >Cancel</button>
      <button
        on:click|preventDefault={submit}
        type="submit"
        class="inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >Sync</button>
    </div>
  </div>
</div>
