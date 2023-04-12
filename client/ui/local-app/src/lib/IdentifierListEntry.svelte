<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { IdentifierEntry } from '~/types';

  export let identifier: IdentifierEntry;

  // component event dispatcher
  // will be used to dispatch the events:
  //  - show_details: show detail for a selected identifier
  //  - link_wallet: link a wallet to a selected identifier
  const dispatch = createEventDispatcher();

  // format a date value to a common textual representation
  // using the user's locale.
  function formatDate(val: string): string {
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
</script>

<tr
  on:click|preventDefault|stopPropagation={() => {
    dispatch('show_details', { did: identifier });
  }}
  class="cursor-pointer odd:bg-white even:bg-slate-50 hover:bg-slate-100">
  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-3">
    {identifier.name}
  </td>
  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
    {identifier.did}
  </td>
  <td class="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
    {identifier.addresses.length}
  </td>
  <td class="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
    {formatDate(identifier.last_sync)}
  </td>
  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
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
  </td>
  <td>
    <button
      on:click|preventDefault|stopPropagation={() => {
        dispatch('link_wallet', { did: identifier });
      }}
      type="button"
      class="inline-flex items-center px-2.5 py-1.5 text-sm font-semibold text-gray-900 hover:text-indigo-600 disabled:cursor-not-allowed disabled:opacity-30">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke-width="1.5"
        stroke="currentColor"
        class="h-5 w-5">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
      </svg>
    </button>
  </td>
</tr>
