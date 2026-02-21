<script lang="ts">
	let {
		class: className = '',
		children,
		...restProps
	}: {
		class?: string;
		children?: import('svelte').Snippet;
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		[key: string]: any;
	} = $props();
</script>

<label
	class="block text-sm font-medium text-gray-700 dark:text-gray-200 {className}"
	{...restProps}
>
	{#if children}{@render children()}{/if}
</label>
