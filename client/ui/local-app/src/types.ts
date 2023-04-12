import type { SvelteComponent } from 'svelte';

/**
 * main application state
 */
export interface AppState {
  /**
   * connected wallet details
   */
  wallet: WalletMetadata;

  /**
   * DIDs available
   */
  identifiers?: IdentifierEntry[];
}

/**
 * connected wallet details
 */
export interface WalletMetadata {
  /**
   * whether the wallet is connected or not
   */
  connected: boolean;

  /**
   * addresses enabled by the user
   */
  addresses: string[];

  /**
   * application name
   */
  name: string;

  /**
   * application description
   */
  description: string;

  /**
   * application home URL
   */
  url: string;

  /**
   * custom application icons
   */
  icons: string[];
}

/**
 * existing managed identifiers
 */
export interface IdentifierEntry {
  /**
   * DID local reference name
   */
  name: string;

  /**
   * actual DID
   */
  did: string;

  /**
   * ALGO addresses linked to the identifier
   */
  addresses: AddressEntry[];

  /**
   * whether the DID remains active
   */
  active: boolean;

  /**
   * date when the DID was last sync to the network,
   * if at all
   */
  last_sync?: string;

  /**
   * raw DID document
   */
  document: unknown;
}

/**
 * settings available when presenting a modal window
 */
export interface ModalOptions {
  /**
   * modal main title
   */
  title: string;

  /**
   * secondary title; optional
   */
  subtitle?: string;

  /**
   * whether to show the modal as a side panel
   */
  asPanel: boolean;

  /**
   * child component to render
   */
  content: typeof SvelteComponent;

  /**
   * properties to pass to the child component
   */
  props?: unknown;
}

/**
 * ALGO address available to a DID
 */
export interface AddressEntry {
  /**
   * ALGO address in text format.
   */
  address: string;

  /**
   * Algorand network the address belongs to.
   */
  network: string;

  /**
   * whether the address is enabled/linked on the DID or not
   */
  enabled: boolean;
}

/**
 * alert status code
 */
export type AlertStatus = 'success' | 'warning' | 'error';
