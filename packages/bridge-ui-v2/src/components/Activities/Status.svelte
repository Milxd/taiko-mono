<script lang="ts">
  import { type Address, fetchBalance, switchNetwork } from '@wagmi/core';
  import { onDestroy, onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { parseEther, UserRejectedRequestError } from 'viem';

  import {
    errorToast,
    infoToast,
    successToast,
    warningToast,
  } from '$components/NotificationToast/NotificationToast.svelte';
  import { Spinner } from '$components/Spinner';
  import { StatusDot } from '$components/StatusDot';
  import { statusComponent } from '$config';
  import { bridges, type BridgeTransaction, MessageStatus } from '$libs/bridge';
  import { isTransactionProcessable } from '$libs/bridge/isTransactionProcessable';
  import { PollingEvent, startPolling } from '$libs/bridge/messageStatusPoller';
  import { chainUrlMap } from '$libs/chain';
  import {
    InsufficientBalanceError,
    InvalidProofError,
    NotConnectedError,
    ProcessMessageError,
    ReleaseError,
    RetryError,
  } from '$libs/error';
  import { getConnectedWallet } from '$libs/util/getConnectedWallet';
  import { getLogger } from '$libs/util/logger';
  import { account } from '$stores/account';
  import { network } from '$stores/network';
  import { pendingTransactions } from '$stores/pendingTransactions';

  const log = getLogger('components:Status');

  export let bridgeTx: BridgeTransaction;

  let polling: ReturnType<typeof startPolling>;

  // UI state
  let processable = false; // bridge tx state to be processed: claimed/retried/released
  let bridgeTxStatus: Maybe<MessageStatus>;

  // TODO: enum?
  let loading: 'claiming' | 'releasing' | false = false;

  function onProcessable(isTxProcessable: boolean) {
    processable = isTxProcessable;
  }

  function onStatusChange(status: MessageStatus) {
    // Keeping model and UI in sync
    bridgeTxStatus = bridgeTx.status = status;
  }

  async function ensureCorrectChain(currentChainId: number, wannaBeChainId: number) {
    const isCorrectChain = currentChainId === wannaBeChainId;
    log(`Are we on the correct chain? ${isCorrectChain}`);

    if (!isCorrectChain) {
      // TODO: shouldn't we inform the user about this change? wallet will popup,
      //       but it's not clear why
      await switchNetwork({ chainId: wannaBeChainId });
    }
  }

  async function checkEnoughBalance(address: Address) {
    const balance = await fetchBalance({ address });
    if (balance.value < parseEther(String(statusComponent.minimumEthToClaim))) {
      throw new InsufficientBalanceError('user has insufficient balance');
    }
  }

  async function claim() {
    if (!$network || !$account?.address) return;

    loading = 'claiming';

    try {
      const { msgHash, message } = bridgeTx;

      if (!msgHash || !message) {
        throw new Error('Missing msgHash or message');
      }

      // Step 1: make sure the user is on the correct chain
      await ensureCorrectChain(Number($network.id), Number(bridgeTx.destChainId));

      // Step 2: make sure the user has enough balance on the destination chain
      await checkEnoughBalance($account.address);

      // Step 3: Find out the type of bridge: ETHBridge, ERC20Bridge, etc
      const bridge = bridges[bridgeTx.tokenType];

      // Step 4: get the user's wallet
      // TODO: do we really need to pass the chainId here? we are already on the dest chain
      const wallet = await getConnectedWallet(Number(bridgeTx.destChainId));

      log(`Claiming ${bridgeTx.tokenType} for transaction`, bridgeTx);

      // Step 5: Call claim() method on the bridge
      const txHash = await bridge.claim({ msgHash, message, wallet });

      const { explorerUrl } = chainUrlMap[Number(bridgeTx.destChainId)];

      infoToast(
        $t('activities.actions.claim.tx', {
          values: {
            token: bridgeTx.symbol,
            url: `${explorerUrl}/tx/${txHash}`,
          },
        }),
      );

      await pendingTransactions.add(txHash, Number(bridgeTx.destChainId));

      successToast(
        $t('activities.actions.claim.success', {
          values: {
            network: $network.name,
          },
        }),
      );

      // We trigger this event to manually to update the UI
      onStatusChange(MessageStatus.DONE);
    } catch (err) {
      console.error(err);

      switch (true) {
        case err instanceof NotConnectedError:
          warningToast($t('messages.account.required'));
          break;
        case err instanceof UserRejectedRequestError:
          warningToast($t('activities.actions.claim.rejected'));
          break;
        case err instanceof InsufficientBalanceError:
          errorToast($t('activities.errors.insufficient_balance'));
          break;
        case err instanceof InvalidProofError:
          errorToast($t('TODO: InvalidProofError'));
          break;
        case err instanceof ProcessMessageError:
          errorToast($t('TODO: ProcessMessageError'));
          break;
        case err instanceof RetryError:
          errorToast($t('TODO: RetryError'));
          break;
        default:
          errorToast($t('TODO: UnknownError'));
          break;
      }
    } finally {
      loading = false;
    }
  }

  async function release() {
    if (!$network || !$account?.address) return;

    loading = 'releasing';

    try {
      const { msgHash, message } = bridgeTx;

      if (!msgHash || !message) {
        throw new Error('Missing msgHash or message');
      }

      // Step 1: make sure the user is on the correct chain
      await ensureCorrectChain(Number($network.id), Number(bridgeTx.srcChainId));

      // Step 2: Find out the type of bridge: ETHBridge, ERC20Bridge, etc
      const bridge = bridges[bridgeTx.tokenType];

      // Step 3: get the user's wallet
      // TODO: might not be needed to pass the chainId here
      const wallet = await getConnectedWallet(Number(bridgeTx.srcChainId));

      log(`Releasing ${bridgeTx.tokenType} for transaction`, bridgeTx);

      // Step 4: Call release() method on the bridge
      const txHash = await bridge.claim({ msgHash, message, wallet });

      const { explorerUrl } = chainUrlMap[Number(bridgeTx.srcChainId)];

      infoToast(
        $t('activities.actions.release.tx', {
          values: {
            token: bridgeTx.symbol,
            url: `${explorerUrl}/tx/${txHash}`,
          },
        }),
      );

      await pendingTransactions.add(txHash, Number(bridgeTx.srcChainId));

      successToast(
        $t('activities.actions.release.success', {
          values: {
            network: $network.name,
          },
        }),
      );
    } catch (err) {
      console.error(err);

      switch (true) {
        case err instanceof NotConnectedError:
          warningToast($t('messages.account.required'));
          break;
        case err instanceof UserRejectedRequestError:
          warningToast($t('activities.actions.release_rejected'));
          break;
        case err instanceof InvalidProofError:
          errorToast($t('TODO: InvalidProofError'));
          break;
        case err instanceof ReleaseError:
          errorToast($t('TODO: ReleaseError'));
          break;
        default:
          errorToast($t('TODO: UnknownError'));
          break;
      }
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (bridgeTx) {
      bridgeTxStatus = bridgeTx.status; // get the current status

      // Can we start claiming/retrying/releasing?
      processable = await isTransactionProcessable(bridgeTx);

      try {
        polling = startPolling(bridgeTx);

        // If there is no emitter, means the bridgeTx is already DONE
        // so we do nothing here
        if (polling?.emitter) {
          // The following listeners will trigger change in the UI
          polling.emitter.on(PollingEvent.PROCESSABLE, onProcessable);
          polling.emitter.on(PollingEvent.STATUS, onStatusChange);
        }
      } catch (err) {
        console.error(err);
        // TODO: handle error
      }
    }
  });

  onDestroy(() => {
    if (polling) {
      polling.destroy();
    }
  });
</script>

<div class="Status f-items-center space-x-1">
  {#if !processable}
    <StatusDot type="pending" />
    <span>{$t('activities.status.initiated.name')}</span>
  {:else if loading}
    <div class="f-items-center space-x-2">
      <Spinner />
      <span>{$t(`activities.status.${loading}`)}</span>
    </div>
  {:else if bridgeTxStatus === MessageStatus.NEW}
    <button class="status-btn" on:click={claim}>
      {$t('activities.button.claim')}
    </button>
  {:else if bridgeTxStatus === MessageStatus.RETRIABLE}
    <button class="status-btn" on:click={claim}>
      {$t('activities.button.claim')}
    </button>
  {:else if bridgeTxStatus === MessageStatus.DONE}
    <StatusDot type="success" />
    <span>{$t('activities.status.claimed.name')}</span>
  {:else if bridgeTxStatus === MessageStatus.FAILED}
    <button class="status-btn" on:click={release}>
      {$t('activities.button.claim')}
    </button>
  {:else}
    <!-- TODO: look into this possible state -->
    <StatusDot type="error" />
    <span>{$t('activities.status.error.name')}</span>
  {/if}
</div>
