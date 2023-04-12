import { getContext, setContext } from 'svelte';
import type { AlertStatus, ModalOptions } from '~/types';

/**
 * application context interface
 */
export type AppContext = {
  /**
   * disply an alert/notification message
   */
  showAlert(status: AlertStatus, message: string): void;

  /**
   * present a modal window
   */
  showModal(options: ModalOptions): void;

  /**
   * close the modal window
   */
  closeModal(): void;
};

/**
 * unique key for application context instance
 */
const contextKey = Symbol('app-context-key');

/**
 * set application context
 */
export const SetContext = function (ctx: AppContext) {
  setContext(contextKey, ctx);
};

/**
 * get main application context instance
 */
export const GetContext = function (): AppContext {
  return getContext(contextKey);
};
