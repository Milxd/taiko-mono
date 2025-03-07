<script lang="ts" context="module">
  export const drawerToggleId = 'side-drawer-toggle';
</script>

<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';

  import { page } from '$app/stores';
  import { Icon } from '$components/Icon';
  import { LinkButton } from '$components/LinkButton';
  import { LogoWithText } from '$components/Logo';
  import { PUBLIC_GUIDE_URL, PUBLIC_L2_EXPLORER_URL } from '$env/static/public';

  let darkTheme: boolean;
  let drawerToggleElem: HTMLInputElement;

  function closeDrawer() {
    drawerToggleElem.checked = false;
  }

  function onMenuKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' || event.key === 'Enter') {
      closeDrawer();
    }
  }

  function getIconFillClass(active: boolean) {
    return active ? 'fill-white' : 'fill-primary-icon';
  }

  function switchTheme(dark: boolean) {
    if (typeof darkTheme === 'undefined') return;

    const theme = dark ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }

  $: isBridgePage = $page.route.id === '/' || $page.route.id === '/nft';
  $: isFaucetPage = $page.route.id === '/faucet';
  $: isActivitiesPage = $page.route.id === '/activities';

  $: switchTheme(darkTheme);

  onMount(() => {
    darkTheme = localStorage.getItem('theme')?.toLocaleLowerCase() === 'dark';
  });
</script>

<div class="drawer md:drawer-open">
  <input id={drawerToggleId} type="checkbox" class="drawer-toggle" bind:this={drawerToggleElem} />

  <div class="drawer-content relative">
    <slot />
  </div>

  <!-- Side drawer's z-index (20) must be greater than content's header (10)-->
  <div class="drawer-side z-20">
    <label for={drawerToggleId} class="drawer-overlay" />

    <!--
      Slow transitions can be pretty annoying after a while.
      Let's reduce it to 100ms for a better experience.
    -->
    <div class="w-h-full !duration-100">
      <header class="flex justify-end py-[20px] px-4 md:hidden">
        <button on:click={closeDrawer} class="h-9">
          <Icon type="x-close" fillClass="fill-primary-icon" size={24} />
        </button>
      </header>

      <aside
        class="
        h-full
        px-[20px]
        md:mt-0
        md:px-4
        md:py-8
        md:w-[226px]
      ">
        <a href="/" class="hidden md:inline-block">
          <LogoWithText textFillClass="fill-primary-content" />
        </a>

        <div role="button" tabindex="0" on:click={closeDrawer} on:keydown={onMenuKeydown}>
          <ul class="menu p-0 md:pt-10 space-y-2">
            <li>
              <LinkButton active={isBridgePage}>
                <Icon type="bridge" fillClass={getIconFillClass(isBridgePage)} />
                <span>{$t('nav.bridge')}</span>
              </LinkButton>
            </li>
            <li>
              <LinkButton href="/faucet" active={isFaucetPage}>
                <Icon type="faucet" fillClass={getIconFillClass(isFaucetPage)} />
                <span>{$t('nav.faucet')}</span>
              </LinkButton>
            </li>
            <li>
              <LinkButton href="/activities" active={isActivitiesPage}>
                <Icon type="activities" fillClass={getIconFillClass(isActivitiesPage)} />
                <span>{$t('nav.activities')}</span>
              </LinkButton>
            </li>
            <li class="border-t border-t-divider-border pt-2">
              <LinkButton href={PUBLIC_L2_EXPLORER_URL} external>
                <Icon type="explorer" />
                <span>{$t('nav.explorer')}</span>
              </LinkButton>
            </li>
            <li>
              <LinkButton href={PUBLIC_GUIDE_URL} external>
                <Icon type="guide" />
                <span>{$t('nav.guide')}</span>
              </LinkButton>
            </li>
            <li>
              <label>
                <Icon type="adjustments" />
                <span>{$t('nav.theme')}</span>
                <input type="checkbox" class="toggle" bind:checked={darkTheme} />
              </label>
            </li>
          </ul>
        </div>
      </aside>
    </div>
  </div>
</div>
