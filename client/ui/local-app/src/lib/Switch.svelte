<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';

  // component event dispatcher
  // will be used to dispatch the events:
  //  - toggle: switch state changed
  const dispatch = createEventDispatcher();

  // component properties
  // screen reader description
  export let description = '';
  // initial state
  export let enabled = false;

  onMount(() => {
    adjustStyles();
  });

  let marker: HTMLElement;
  let bg: HTMLElement;

  // alternate the switch state
  export function toggle(): void {
    enabled = !enabled;
    dispatch('toggle', { enabled });
    adjustStyles();
  }

  // return the current state of the switch
  export function isEnabled(): boolean {
    return enabled;
  }

  function adjustStyles(): void {
    // mark as 'enabled'
    if (enabled) {
      bg.classList.add('bg-indigo-600');
      bg.classList.remove('bg-gray-200');
      marker.classList.add('translate-x-5');
      return;
    }
    // mark as 'disabled'
    bg.classList.add('bg-gray-200');
    bg.classList.remove('bg-indigo-600');
    marker.classList.remove('translate-x-5');
  }
</script>

<button
  bind:this={bg}
  on:click|preventDefault={toggle}
  type="button"
  class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent bg-gray-200 transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2"
  role="switch"
  aria-checked="false">
  {#if description}
    <span class="sr-only">{description}</span>
  {/if}
  <span
    bind:this={marker}
    aria-hidden="true"
    class="pointer-events-none inline-block h-5 w-5 translate-x-0 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out" />
</button>
