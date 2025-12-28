import { createStorageAPI } from '$lib/api/base';
import type { Pattern, PatternDescription } from '$lib/interfaces/pattern-interface';
import { get, writable, derived } from 'svelte/store';
import { languageStore } from './language-store';

// Store for all patterns
const allPatterns = writable<Pattern[]>([]);

// Filtered patterns based on language
export const patterns = derived([allPatterns, languageStore], ([$allPatterns, $language]) => {
	if (!$language) return $allPatterns;
	// If language is selected, filter out patterns of other languages
	return $allPatterns.filter((p) => {
		// Keep all patterns if no language is selected
		if (!$language) return true;

		// Check if pattern has a language prefix (e.g., en_, fr_)
		const match = p.Name.match(/^([a-z]{2})_/);
		if (!match) return true; // Keep patterns without language prefix

		// Only filter out patterns that have a different language prefix
		const patternLang = match[1];
		return patternLang === $language;
	});
});

export const systemPrompt = writable<string>('');
export const selectedPatternName = writable<string>('');

// Pattern variables store
export const patternVariables = writable<Record<string, string>>({});

export const setSystemPrompt = (prompt: string) => {
	systemPrompt.set(prompt);
};

export const patternAPI = {
	...createStorageAPI<Pattern>('patterns'),

	async loadPatterns() {
		try {
			// First load pattern descriptions
			const descriptionsResponse = await fetch('/data/pattern_descriptions.json');
			const descriptionsData = await descriptionsResponse.json();
			const descriptions = descriptionsData.patterns as PatternDescription[];

			// Then load pattern names and contents
			const response = await fetch(`/api/patterns/names`);
			const data = await response.json();

			// Create an array of promises to fetch all pattern contents
			const patternsPromises = data.map(async (pattern: string) => {
				try {
					const patternResponse = await fetch(`/api/patterns/${pattern}`);
					const patternData = await patternResponse.json();

					// Find matching description from JSON
					const desc = descriptions.find((d) => d.patternName === pattern);
					if (!desc) {
						console.warn(`No description found for pattern: ${pattern}`);
					}

					return {
						Name: pattern,
						Description: desc?.description || pattern.charAt(0).toUpperCase() + pattern.slice(1),
						Pattern: patternData.Pattern || '',
						tags: desc?.tags || [] // Add tags from description
					};
				} catch (error) {
					console.error(`Failed to load pattern ${pattern}:`, error);
					// Still try to get description even if pattern content fails
					const desc = descriptions.find((d) => d.patternName === pattern);
					return {
						Name: pattern,
						Description: desc?.description || pattern.charAt(0).toUpperCase() + pattern.slice(1),
						Pattern: '',
						tags: desc?.tags || [] // Add tags here too for consistency
					};
				}
			});

			// Wait for all pattern contents to be fetched
			const loadedPatterns = await Promise.all(patternsPromises);
			allPatterns.set(loadedPatterns);
			return loadedPatterns;
		} catch (error) {
			console.error('Failed to load patterns:', error);
			allPatterns.set([]);
			return [];
		}
	},

	selectPattern(patternName: string) {
		const patterns = get(allPatterns);
		const selectedPattern = patterns.find((p) => p.Name === patternName);
		if (selectedPattern) {
			setSystemPrompt(selectedPattern.Pattern);
			selectedPatternName.set(patternName); // Make sure this is set before setting system prompt
		} else {
			setSystemPrompt('');
			selectedPatternName.set('');
		}
	}
};
