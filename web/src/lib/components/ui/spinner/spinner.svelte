<script lang="ts">
  import { cn } from '$lib/utils/utils';
  import { Loader2 } from 'lucide-svelte';

  let { class: className = undefined, ...restProps } = $props();
</script>

<Loader2 class={cn('h-4 w-4 animate-spin', className)} {...restProps} />
