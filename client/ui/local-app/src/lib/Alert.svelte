<script lang="ts">
  import { slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import type { AlertStatus } from '~/types';

  // alert is hidden by default
  let hidden = true;

  // component properties
  let message: string;
  let status: AlertStatus = 'success';

  // show the alert message.
  // this function can be called by the component's parents.
  export function show(kind: AlertStatus, msg: string): void {
    status = kind;
    message = msg;
    hidden = false;
  }

  // hide the alert message.
  // this function can be called by the component's parents.
  export function close(): void {
    hidden = true;
  }

  // properly style the alert message based on its status
  function styles(section: 'button' | 'text' | 'border'): string {
    switch (section) {
      case 'border':
        return 'border-COLOR-400 bg-COLOR-50'.replaceAll('COLOR', statusColor());
      case 'text':
        return 'text-COLOR-800'.replaceAll('COLOR', statusColor());
      case 'button':
        return 'bg-COLOR-50 text-COLOR-500 hover:bg-COLOR-100 focus:ring-COLOR-600 focus:ring-offset-COLOR-50'.replaceAll(
          'COLOR',
          statusColor()
        );
    }
  }

  // map alert status to a color
  function statusColor(): string {
    switch (status) {
      case 'success':
        return 'green';
      case 'warning':
        return 'yellow';
      case 'error':
        return 'red';
    }
  }
</script>

{#if !hidden}
  <div
    class="border-l-4 p-4 {styles('border')}"
    transition:slide={{ easing: cubicOut, duration: 400 }}>
    <div class="flex">
      <div class="ml-3">
        <p class="text-sm font-medium {styles('text')}">{message}</p>
      </div>
      <div class="ml-auto pl-3">
        <div class="-mx-1.5 -my-1.5">
          <button
            on:click|preventDefault={() => {
              close();
            }}
            type="button"
            class="inline-fle p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2 {styles(
              'button'
            )}">
            <span class="sr-only">Dismiss</span>
            <svg
              class="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden="true">
              <path
                d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!--
      invisible elements required to force the tailwind compiler
      to include all the different styles used dynamically by
      this component.
    -->
    <div class="hidden">
      <p
        class="border-green-400 bg-green-50 text-green-800 hover:bg-green-100 focus:ring-green-600 focus:ring-offset-green-50">
        <span class="text-green-500">success styles</span>
      </p>
      <p
        class="border-red-400 bg-red-50 text-red-800 hover:bg-red-100 focus:ring-red-600 focus:ring-offset-red-50">
        <span class="text-red-500">error styles</span>
      </p>
      <p
        class="border-yellow-400 bg-yellow-50 text-yellow-800 hover:bg-yellow-100 focus:ring-yellow-600 focus:ring-offset-yellow-50">
        <span class="text-yellow-500">warning styles</span>
      </p>
    </div>
  </div>
{/if}
