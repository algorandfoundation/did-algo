<script lang="ts">
  import Connector from '~/lib/Connector.svelte';
  import IdentifierList from '~/lib/IdentifierList.svelte';
  import Alert from '~/lib/Alert.svelte';
  import Modal from '~/lib/Modal.svelte';
  import { appState } from '~/store';
  import { SetContext } from '~/context';
  import type { AlertStatus, ModalOptions } from '~/types';
  import type { SvelteComponent } from 'svelte';

  let alertMsg: Alert;
  let modal: Modal;
  let modalContent: typeof SvelteComponent;
  let modalContentProps: object;

  // load state during app initialization
  appState.reload();

  // define main application context interface
  SetContext({
    showAlert(st: AlertStatus, msg: string): void {
      alertMsg.show(st, msg);
    },
    showModal(options: ModalOptions): void {
      modalContent = options.content;
      modalContentProps = options.props as object;
      modal.show(options.title, options.subtitle, options.asPanel);
    },
    closeModal(): void {
      modal.close();
    }
  });
</script>

<!-- modal window -->
<Modal bind:this={modal}>
  <svelte:component this={modalContent} {...modalContentProps} />
</Modal>

<!-- content wrapper -->
<div class="relative flex flex-col justify-center bg-gray-50 py-6 sm:py-12">
  <!-- bg image -->
  <img
    src="/img/beams.jpg"
    alt=""
    class="absolute left-1/2 top-1/2 max-w-none -translate-x-1/2 -translate-y-1/2"
    width="1308" />
  <div
    class="absolute inset-0 bg-[url(/img/grid.svg)] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))]" />

  <!-- content -->
  <div
    class="relative bg-white px-6 pb-8 pt-10 shadow-xl ring-1 ring-gray-900/5 sm:mx-auto sm:px-10 md:w-3/4 md:rounded-md">
    <div class="space-y-6 py-6 text-base leading-7 text-gray-600">
      <!-- header -->
      <div class="relative w-full">
        <h1 class="inline-block text-2xl text-gray-800">AlgoID Connect</h1>
        <Connector mainnet={true} />
      </div>

      <!-- desc -->
      <p>
        Use this graphical interface to manage your <code class="font-mono text-pink-400"
          >did:algo</code>
        Decentralized Identifiers. Connect your wallet and link your
        <code class="font-bold">ALGO</code> addresses to a DID to enable account discovery
        and facilitate payments and other interactions.
      </p>

      <!-- alerts and notifications -->
      <Alert bind:this={alertMsg} />

      <!-- content -->
      <IdentifierList identifiers={$appState.identifiers} />
    </div>
  </div>

  <!-- footer -->
  <div class="relative mx-auto mt-6 w-1/2 text-center">
    <p class="text-sm text-gray-400">
      For more information, or to get the source code for this application, checkout the <a
        target="_blank"
        href="http://github.com/algorandfoundation/did-algo"
        class="text-indigo-600">official repository</a
      >.
    </p>
  </div>
</div>
