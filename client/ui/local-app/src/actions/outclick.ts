import type { ActionReturn } from 'svelte/action';

interface Attributes {
  // define the custom attribute the will be added on the HTMLElement
  // that uses this action. This fixes typescript errors.
  'on:outclick': () => void;
}

export function outclick(node: HTMLElement): ActionReturn<unknown, Attributes> {
  const handleClick = (event: MouseEvent) => {
    if (!node.contains(event.target as HTMLElement)) {
      node.dispatchEvent(new CustomEvent('outclick'));
    }
  };
  document.addEventListener('click', handleClick, true);

  return {
    destroy() {
      document.removeEventListener('click', handleClick, true);
    }
  };
}
