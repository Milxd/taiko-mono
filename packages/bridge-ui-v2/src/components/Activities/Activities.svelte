<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import type { Address } from 'viem';

  import { Card } from '$components/Card';
  import ChainSelector from '$components/ChainSelector/ChainSelector.svelte';
  import { DesktopOrLarger } from '$components/DesktopOrLarger';
  import OnAccount from '$components/OnAccount/OnAccount.svelte';
  import { Paginator } from '$components/Paginator';
  import { Spinner } from '$components/Spinner';
  import { activitiesConfig } from '$config';
  import { type BridgeTransaction, fetchTransactions } from '$libs/bridge';
  import { bridgeTxService } from '$libs/storage';
  import { account, network } from '$stores';
  import type { Account } from '$stores/account';

  import MobileDetailsDialog from './MobileDetailsDialog.svelte';
  import StatusInfoDialog from './StatusInfoDialog.svelte';
  import Transaction from './Transaction.svelte';

  let transactions: BridgeTransaction[] = [];

  let currentPage = 1;

  let isBlurred = false;
  const transitionTime = activitiesConfig.blurTransitionTime;

  let totalItems = 0;
  let pageSize = activitiesConfig.pageSizeDesktop;

  let loadingTxs = false;

  let detailsOpen = false;
  let isDesktopOrLarger: boolean;

  let selectedItem: BridgeTransaction | null = null;

  const handlePageChange = (detail: number) => {
    isBlurred = true;
    setTimeout(() => {
      currentPage = detail;
      isBlurred = false;
    }, transitionTime);
  };

  const getTransactionsToShow = (page: number, pageSize: number, bridgeTx: BridgeTransaction[]) => {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return bridgeTx.slice(start, end);
  };

  const onAccountChange = async (newAccount: Account, oldAccount?: Account) => {
    // We want to make sure that we are connected and only
    // fetch if the account has changed
    if (newAccount?.isConnected && newAccount.address && newAccount.address !== oldAccount?.address) {
      loadingTxs = true;

      try {
        updateTransactions(newAccount.address);
      } catch (err) {
        console.error(err);
        // TODO: handle
      } finally {
        loadingTxs = false;
      }
    }
  };

  const closeDetails = () => {
    detailsOpen = false;
    selectedItem = null;
  };

  const openDetails = (tx: BridgeTransaction) => {
    detailsOpen = true;
    selectedItem = tx;
  };

  const updateTransactions = async (address: Address) => {
    const { mergedTransactions, outdatedLocalTransactions } = await fetchTransactions(address);
    transactions = mergedTransactions;
    if (outdatedLocalTransactions.length > 0) {
      await bridgeTxService.removeTransactions(address, outdatedLocalTransactions);
    }
  };

  onMount(async () => {
    // We want to make sure that we are connected before fetching
    if (!$account?.isConnected || !$account?.address) return;

    loadingTxs = true;

    try {
      await updateTransactions($account.address);
    } catch (err) {
      console.error(err);
      // TODO: handle
    } finally {
      loadingTxs = false;
    }
  });

  $: pageSize = isDesktopOrLarger ? activitiesConfig.pageSizeDesktop : activitiesConfig.pageSizeMobile;

  $: transactionsToShow = getTransactionsToShow(currentPage, pageSize, transactions);

  $: totalItems = transactions.length;

  // Some shortcuts to make the code more readable
  $: isConnected = $account?.isConnected;
  $: hasTxs = transactions.length > 0;

  // Controls what we render on the page
  $: renderLoading = loadingTxs && isConnected;
  $: renderTransactions = !renderLoading && isConnected && hasTxs;
  $: renderNoTransactions = !renderLoading && !renderTransactions;
</script>

<div class="flex flex-col justify-center w-full">
  <Card title={$t('activities.title')} text={$t('activities.description')}>
    <div class="space-y-[35px]">
      <ChainSelector label={$t('chain_selector.currently_on')} value={$network} switchWallet small />
      <div class="flex flex-col" style={`min-height: calc(${transactionsToShow.length} * 80px);`}>
        {#if isDesktopOrLarger}
          <div class="h-sep" />
          <div class="text-primary-content flex">
            <div class="w-1/5 py-2">{$t('activities.header.from')}</div>
            <div class="w-1/5 py-2">{$t('activities.header.to')}</div>
            <div class="w-1/5 py-2">{$t('activities.header.amount')}</div>
            <div class="w-1/5 py-2 flex flex-row">
              {$t('activities.header.status')}
              <StatusInfoDialog />
            </div>
            <div class="w-1/5 py-2">{$t('activities.header.explorer')}</div>
          </div>
          <div class="h-sep" />
        {/if}

        {#if renderLoading}
          <div class="flex items-center justify-center text-primary-content h-[80px]">
            <Spinner /> <span class="pl-3">{$t('common.loading')}...</span>
          </div>
        {/if}

        {#if renderTransactions}
          <div
            class="flex flex-col items-center"
            style={isBlurred ? `filter: blur(5px); transition: filter ${transitionTime / 1000}s ease-in-out` : ''}>
            {#each transactionsToShow as item (item.hash)}
              <Transaction {item} on:click={isDesktopOrLarger ? undefined : () => openDetails(item)} />
              <div class="h-sep" />
            {/each}
          </div>
        {/if}

        {#if renderNoTransactions}
          <div class="flex items-center justify-center text-primary-content h-[80px]">
            <span class="pl-3">{$t('activities.no_transactions')}</span>
          </div>
        {/if}
      </div>
    </div>
  </Card>

  <div class="flex justify-end pt-2">
    <Paginator {pageSize} {totalItems} on:pageChange={({ detail }) => handlePageChange(detail)} />
  </div>
</div>

<MobileDetailsDialog {closeDetails} {detailsOpen} {selectedItem} />

<OnAccount change={onAccountChange} />

<DesktopOrLarger bind:is={isDesktopOrLarger} />
