import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { YoutubeTranscript } from 'youtube-transcript';

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json();

		// Handle YouTube URL request
		if (body.url) {
			// Extract video ID
			const match = body.url.match(
				/(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})/
			);
			const videoId = match ? match[1] : null;

			if (!videoId) {
				return json({ error: 'Invalid YouTube URL' }, { status: 400 });
			}

			const transcriptItems = await YoutubeTranscript.fetchTranscript(videoId);
			const transcript = transcriptItems.map((item) => item.text).join(' ');

			// Create response with transcript and language
			const response = {
				transcript,
				title: videoId,
				language: body.language
			};

			return json(response);
		}

		// Handle pattern execution request
		const fabricResponse = await fetch('http://localhost:8080/api/chat', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(body)
		});

		if (!fabricResponse.ok) {
			console.error('Error from Fabric API:', {
				status: fabricResponse.status,
				statusText: fabricResponse.statusText
			});
			throw new Error(`Fabric API error: ${fabricResponse.statusText}`);
		}

		const stream = fabricResponse.body;
		if (!stream) {
			throw new Error('No response from fabric backend');
		}

		// Return the stream
		const response = new Response(stream, {
			headers: {
				'Content-Type': 'text/event-stream',
				'Cache-Control': 'no-cache',
				Connection: 'keep-alive'
			}
		});

		return response;
	} catch (error) {
		console.error('\n=== Error ===');
		console.error('Type:', error?.constructor?.name);
		console.error('Message:', error instanceof Error ? error.message : String(error));
		console.error('Stack:', error instanceof Error ? error.stack : 'No stack trace');
		return json(
			{ error: error instanceof Error ? error.message : 'Failed to process request' },
			{ status: 500 }
		);
	}
};
