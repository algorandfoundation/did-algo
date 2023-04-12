<script lang="ts" strictEvents>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { outclick } from '~/actions/outclick';

  // component internal state
  let state = {
    hidden: true, // modal start hidden by default
    asPanel: false, // display as modal by default
    title: '',
    subtitle: ''
  };

  // dispatch custom component event
  //  - close: the modal was closed
  //  - open: the modal was opened
  const dispatch = createEventDispatcher();

  // whether the modal is currently visible
  export function isVisible(): boolean {
    return !state.hidden;
  }

  // show the modal window.
  // this function can be called by the component's parents.
  export function show(title: string, subtitle: string, asPanel: boolean): void {
    state = {
      title,
      subtitle,
      asPanel,
      hidden: false
    };
  }

  // hide the modal window.
  // this function can be called by the component's parents.
  export function close(): void {
    state = { ...state, hidden: true };
  }

  // custom transition to reveal the modal contents
  function reveal(node: HTMLElement, params?: unknown) {
    params = {
      x: state.asPanel ? 100 : 0, // panel slide from right
      y: state.asPanel ? 0 : 100, // modal slide from bottom
      duration: 400,
      easing: cubicOut
    };
    return fly(node, params);
  }

  // adjust component styles based on the `asPanel` property
  function styles(section: 'wrapper' | 'content'): string {
    if (section == 'wrapper') {
      // panel appears from right, modal is centered
      return state.asPanel ? 'items-end' : 'items-center justify-center';
    }
    // panel is full height and width between 50% and 75%
    // modal's width is between 75% and 100%; height is based on content
    return state.asPanel
      ? 'h-full w-3/4  md:w-1/2'
      : 'min-h-fit max-h-screen w-full sm:w-3/4 sm:rounded-md lg:w-1/2';
  }

  // listen for keyboard events and close the modal
  // when the user press the `Esc` key.
  function keyDown(event: KeyboardEvent) {
    if (!state.hidden && event.key == 'Escape') {
      close();
    }
  }
</script>

<svelte:window on:keydown={keyDown} />

{#if !state.hidden}
  <div
    transition:fade={{ duration: 300, easing: cubicOut }}
    on:introend={() => {
      dispatch('open');
    }}
    on:outroend={() => {
      dispatch('close');
    }}
    class="absolute bottom-0 left-0 right-0 top-0 z-10 flex flex-col bg-gray-700 bg-opacity-75 backdrop-blur-sm transition-all {styles(
      'wrapper'
    )}">
    <div
      transition:reveal
      use:outclick
      on:outclick={close}
      class="overflow-y-scroll bg-white shadow-2xl {styles('content')}">
      <div class="space-y-4">
        <!-- header -->
        <div class="flex items-center bg-gray-50 p-6 sm:rounded-md">
          <div class="flex-1">
            <h1 class="text-lg font-semibold text-gray-900">{state.title}</h1>
            {#if state.subtitle}
              <span class="text-sm text-gray-500">{state.subtitle}</span>
            {/if}
          </div>
          <button
            on:click={close}
            class="rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2">
            <span class="sr-only">Close</span>
            <svg
              class="h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
              aria-hidden="true">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <!-- content -->
        <slot />
      </div>
    </div>
  </div>
{/if}
