<script lang="ts">
  import IdentifierListEntry from '~/lib/IdentifierListEntry.svelte';
  import IdentifierDetails from '~/lib/IdentifierDetails.svelte';
  import NewIdentifier from '~/lib/NewIdentifier.svelte';
  import ManageAddresses from '~/lib/ManageAddresses.svelte';
  import { GetContext } from '~/context';
  import type { IdentifierEntry } from '~/types';

  const appCtx = GetContext();

  // list of available identifiers
  export let identifiers: IdentifierEntry[];

  // display the 'create new identifier' form
  function createNew() {
    appCtx.showModal({
      title: 'Create New Identifier',
      asPanel: false,
      content: NewIdentifier,
      props: {}
    });
  }

  // display the details for a selected identifier
  function showDetails(event: CustomEvent<{ did: IdentifierEntry }>) {
    appCtx.showModal({
      title: event.detail.did.did,
      subtitle: 'Identifier Details',
      asPanel: true,
      content: IdentifierDetails,
      props: {
        identifier: event.detail.did
      }
    });
  }

  // display the address configuration modal
  function linkWallet(event: CustomEvent<{ did: IdentifierEntry }>) {
    appCtx.showModal({
      title: 'Manage Link Addresses',
      subtitle: event.detail.did.did,
      asPanel: false,
      content: ManageAddresses,
      props: {
        identifier: event.detail.did
      }
    });
  }
</script>

<section>
  {#if identifiers.length == 0}
    <!-- empty state-->
    <button
      on:click={createNew}
      type="button"
      class="relative block w-full rounded-lg border-2 border-dashed border-gray-300 p-12 text-center hover:border-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
      <svg
        class="mx-auto h-12 w-12 text-gray-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5zm6-10.125a1.875 1.875 0 11-3.75 0 1.875 1.875 0 013.75 0zm1.294 6.336a6.721 6.721 0 01-3.17.789 6.721 6.721 0 01-3.168-.789 3.376 3.376 0 016.338 0z" />
      </svg>
      <span class="mt-2 block text-sm font-semibold text-gray-900">
        Create your first decentralized identifier
      </span>
    </button>
  {:else}
    <!-- entries list -->
    <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
      <table class="min-w-full divide-y divide-gray-300">
        <thead>
          <tr>
            <th
              scope="col"
              class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-3">
              Reference
            </th>
            <th
              scope="col"
              class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
              DID
            </th>
            <th
              scope="col"
              class="hidden px-3 py-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell">
              Addresses
            </th>
            <th
              scope="col"
              class="hidden px-3 py-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell">
              Last Sync
            </th>
            <th
              scope="col"
              class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
              Status
            </th>
            <th
              scope="col"
              class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
              <span class="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody class="bg-white">
          {#each identifiers as item}
            <IdentifierListEntry
              identifier={item}
              on:show_details={showDetails}
              on:link_wallet={linkWallet} />
          {/each}
        </tbody>
      </table>
      <!-- permanent 'create new' button -->
      <button
        on:click={createNew}
        type="button"
        class="relative flex w-full flex-row justify-center border-2 bg-gray-700 py-2 text-gray-200 hover:bg-gray-800 hover:text-white">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          class="mr-2 h-8 w-8">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 01-2.25 2.25M16.5 7.5V18a2.25 2.25 0 002.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 002.25 2.25h13.5M6 7.5h3v3H6v-3z" />
        </svg>
        <span class="mt-2 block text-sm font-semibold">
          Create new decentralized identifier
        </span>
      </button>
    </div>
  {/if}
</section>
