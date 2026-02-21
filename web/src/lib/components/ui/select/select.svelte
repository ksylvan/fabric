<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    value = $bindable(undefined),
    disabled = false,
    class: className = undefined,
    children,
    ...restProps
  }: {
    value?: any;
    disabled?: boolean;
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();
</script>

<div class="relative">
	<select
		{disabled}
		bind:value
		class={cn(
			"select flex h-8 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-xs font-medium text-white/90 shadow-lg ring-offset-background placeholder:text-muted-primary focus:outline-none focus:ring-1 focus:ring-ring hover:text-white disabled:cursor-not-allowed disabled:opacity-50 appearance-none transition-colors",
			className
		)}
		{...restProps}
	>
		{#if children}{@render children()}{/if}
	</select>
</div>
