<script lang="ts">
  import Prism from 'svelte-prism';
  import 'prism-themes/themes/prism-atom-dark.css';
  import type { IdentifierEntry } from '~/types';

  export let identifier: IdentifierEntry;

  let didDocument = JSON.stringify(identifier.document, null, 2);

  // format a date value to a common textual representation
  // using the user's locale.
  function formatDate(val: string): string {
    if (val === '') {
      return '-';
    }
    let parsed = new Date();
    parsed.setTime(Date.parse(val));
    return parsed.toLocaleDateString(navigator.language, {
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      month: 'short',
      minute: 'numeric',
      hour12: true
    });
  }

  // format a crypto address to a common textual representation.
  function formatAddress(address: string): string {
    return address.slice(0, 14) + '...' + address.slice(-14);
  }
</script>

<!-- Divider container -->
<div class="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
  <!-- desc -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5 first:sm:pt-0">
    <p class="sm:col-span-3">
      These are the latest details for the selected identifier. You can use the provided
      <span class="inline-block text-indigo-600">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="2"
          stroke="currentColor"
          class="h-4 w-4">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
        </svg>
      </span>
      tool to add or remove <code class="font-bold">ALGO</code> addresses associated with this
      identifier.
    </p>
  </div>
  <!-- name -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    <div>
      <p class="block text-base font-medium leading-6 text-gray-900">
        Local reference (name)
      </p>
    </div>
    <div class="sm:col-span-2">
      <p class="block w-full text-base text-gray-600">{identifier.name}</p>
    </div>
  </div>
  <!-- last sync -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    <div>
      <p class="block text-base font-medium leading-6 text-gray-900">Last sync</p>
    </div>
    <div class="sm:col-span-2">
      <p class="block w-full text-base text-gray-600">
        {formatDate(identifier.last_sync)}
      </p>
    </div>
  </div>
  <!-- linked addresses -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    <div>
      <p class="block text-base font-medium leading-6 text-gray-900">Linked addresses</p>
    </div>
    <div class="sm:col-span-2">
      <p class="block w-full text-base text-gray-600">
        {#if !identifier.addresses.length}
          /
        {:else}
          {#each identifier.addresses as entry}
            <span
              class="inline-flex rounded-full bg-indigo-100 px-2 text-sm font-semibold leading-5 text-indigo-800">
              {formatAddress(entry.address)}
            </span>
            <br />
          {/each}
        {/if}
      </p>
    </div>
  </div>
  <!-- status -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    <div>
      <p class="block text-base font-medium leading-6 text-gray-900">Current status</p>
    </div>
    <div class="sm:col-span-2">
      {#if identifier.active}
        <span
          class="inline-flex rounded-full bg-green-100 px-2 text-xs font-semibold leading-5 text-green-800">
          Active
        </span>
      {:else}
        <span
          class="inline-flex rounded-full bg-red-100 px-2 text-xs font-semibold leading-5 text-red-800">
          Deactivated
        </span>
      {/if}
    </div>
  </div>
  <!-- DID doc -->
  <div
    class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
    <div class="sm:col-span-3">
      <Prism language="javascript" source={didDocument} />
    </div>
  </div>
</div>
