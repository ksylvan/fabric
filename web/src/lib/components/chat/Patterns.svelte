<script lang="ts">
	import { onMount } from 'svelte';
	import { Select } from '$lib/components/ui/select';
	import { patterns, patternAPI, selectedPatternName } from '$lib/store/pattern-store';
	import { get } from 'svelte/store';

	let selectedPreset = $selectedPatternName || '';

	// Subscribe to selectedPatternName changes
	selectedPatternName.subscribe((value) => {
		if (value && value !== selectedPreset) {
			selectedPreset = value;
		}
	});

	// Watch selectedPreset changes
	// Always call selectPattern when the dropdown value changes.
	// The patternAPI.selectPattern function handles empty strings correctly.
	$: {
		try {
			// Call the function to select the pattern (or reset if selectedPreset is empty)
			patternAPI.selectPattern(selectedPreset);
		} catch (error) {
			// Log any errors during the pattern selection process
			console.error('Error processing pattern selection:', error);
		}
	}

	onMount(async () => {
		await patternAPI.loadPatterns();
	});
</script>

<div class="min-w-0">
	<Select
		bind:value={selectedPreset}
		class="bg-primary-800/30 border-none hover:bg-primary-800/40 transition-colors"
	>
		<option value="">Load a pattern...</option>
		{#each $patterns as pattern}
			<option value={pattern.Name}>{pattern.Name}</option>
		{/each}
	</Select>
</div>
