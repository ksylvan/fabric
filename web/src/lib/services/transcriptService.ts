import { languageStore } from '$lib/store/language-store';
import { get } from 'svelte/store';

export interface TranscriptResponse {
	transcript: string;
	title: string;
}

function decodeHtmlEntities(text: string): string {
	const textarea = document.createElement('textarea');
	textarea.innerHTML = text;
	return textarea.value;
}

export async function getTranscript(url: string): Promise<TranscriptResponse> {
	try {
		const originalLanguage = get(languageStore);

		const response = await fetch('/api/youtube/transcript', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				url,
				language: originalLanguage // Pass original language to server
			})
		});

		if (!response.ok) {
			const errorData = await response.json();
			throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
		}

		const data = await response.json();
		if (data.error) {
			throw new Error(data.error);
		}

		// Decode HTML entities in transcript
		data.transcript = decodeHtmlEntities(data.transcript);

		// Ensure language is preserved
		if (get(languageStore) !== originalLanguage) {
			languageStore.set(originalLanguage);
		}

		return data;
	} catch (error) {
		console.error('Transcript fetch error:', error);
		throw error instanceof Error ? error : new Error('Failed to fetch transcript');
	}
}
