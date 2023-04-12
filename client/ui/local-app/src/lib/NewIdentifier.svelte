<script lang="ts">
  import { onMount } from 'svelte';
  import Alert from '~/lib/Alert.svelte';
  import { GetContext } from '~/context';
  import { appState } from '~/store';

  let alertMsg: Alert;
  const appCtx = GetContext();

  onMount(() => {
    document.getElementById('name').focus();
  });

  // validate user input and create a new identifier
  function processForm(event: Event) {
    event.preventDefault();
    const formData = new FormData(event.target as HTMLFormElement);
    const data = Object.fromEntries(formData.entries());
    for (const [key] of Object.entries(data)) {
      if (!validateField(key)) {
        alertMsg.show('error', 'Validate the provided values and try again.');
        return;
      }
    }
    if (data.recovery_key.length < 8) {
      alertMsg.show('error', 'The recovery key must be at least 8 characters long.');
      return;
    }
    if (data.confirmation !== data.recovery_key) {
      alertMsg.show('error', 'The confirmation key does not match the recovery key.');
      return;
    }
    createDID(data.name as string, data.recovery_key as string);
  }

  // validate a form field is not empty
  function validateField(field: string): boolean {
    const target = document.getElementById(field) as HTMLInputElement;
    const value = target.value;
    if (value.length === 0) {
      target.classList.add('text-red-900', 'ring-red-300');
      return false;
    }
    target.classList.remove('text-red-900', 'ring-red-300');
    return true;
  }

  // create a new identifier
  async function createDID(name: string, key: string) {
    let result = await appState.createDID(name, key);
    appCtx.closeModal();
    if (result) {
      appCtx.showAlert('success', 'Identifier created successfully.');
    } else {
      appCtx.showAlert('error', 'Failed to create identifier.');
    }
  }
</script>

<div class="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
  <form on:submit|preventDefault={processForm} class="flex h-full flex-col">
    <!-- desc and alerts -->
    <div
      class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5 first:sm:pt-0">
      <div class="sm:col-span-3">
        <Alert bind:this={alertMsg} />
      </div>
      <p class="sm:col-span-3">
        A decentralized identifier (or DID) is an asset designed to be owned by a <strong
          >controller</strong>
        entity. A single identifier can be used on any number of services, and you can create
        as many identifiers as you want.
      </p>
    </div>
    <!-- name -->
    <div
      class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <p class="block text-base font-medium leading-6 text-gray-900">
          Name (local reference)
        </p>
      </div>
      <div class="sm:col-span-2">
        <input
          type="text"
          name="name"
          id="name"
          class="block w-full rounded-md border-0 p-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" />
      </div>
    </div>
    <!-- recovery key -->
    <div
      class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <p class="sm:col-span-3">
        The recovery key used to create this identifier is not stored locally. If you lose
        it ther's no other way to recover it. <strong
          >Please make sure you have a copy of it.</strong>
      </p>
      <div>
        <p class="block text-base font-medium leading-6 text-gray-900">Recovery key</p>
      </div>
      <div class="sm:col-span-2">
        <input
          type="password"
          name="recovery_key"
          id="recovery_key"
          class="block w-full rounded-md border-0 p-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" />
      </div>
      <div>
        <p class="block text-base font-medium leading-6 text-gray-900">
          Key confirmation
        </p>
      </div>
      <div class="sm:col-span-2">
        <input
          type="password"
          name="confirmation"
          id="confirmation"
          class="block w-full rounded-md border-0 p-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" />
      </div>
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
          type="submit"
          class="inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
          >Create</button>
      </div>
    </div>
  </form>
</div>
