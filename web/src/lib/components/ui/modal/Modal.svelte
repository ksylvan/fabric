<script lang="ts">
	import { fade, scale } from 'svelte/transition';

	let {
		show = false,
		onclose,
		children
	}: {
		show?: boolean;
		onclose?: () => void;
		children?: import('svelte').Snippet;
	} = $props();
</script>

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 mt-2"
		onclick={() => onclose?.()}
		onkeydown={(e) => e.key === 'Escape' && onclose?.()}
		role="dialog"
		aria-modal="true"
		aria-label="Modal dialog"
		tabindex="-1"
		transition:fade={{ duration: 200 }}
	>
		<div
			class="relative"
			onclick={(e) => e.stopPropagation()}
			role="document"
			aria-label="Modal content"
			transition:scale={{ duration: 200 }}
		>
			{#if children}{@render children()}{/if}
		</div>
	</div>
{/if}

<style>
	.fixed {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		left: 0;
	}
</style>
