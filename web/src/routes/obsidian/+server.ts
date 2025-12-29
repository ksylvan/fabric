import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface ObsidianRequest {
	pattern: string;
	noteName: string;
	content: string;
}

function escapeShellArg(arg: string): string {
	// Replace single quotes with '\'' and wrap in single quotes
	return `'${arg.replace(/'/g, "'\\''")}'`;
}

export const POST: RequestHandler = async ({ request }) => {
	let tempFile: string | undefined;

	try {
		// Parse and validate request
		const body = (await request.json()) as ObsidianRequest;
		if (!body.pattern || !body.noteName || !body.content) {
			return json(
				{ error: 'Missing required fields: pattern, noteName, or content' },
				{ status: 400 }
			);
		}

		// Format content with markdown code blocks
		const formattedContent = `\`\`\`markdown\n${body.content}\n\`\`\``;
		const escapedFormattedContent = escapeShellArg(formattedContent);

		// Generate file name and path
		const fileName = `${new Date().toISOString().split('T')[0]}-${body.noteName}.md`;

		const obsidianDir = 'myfiles/Fabric_obsidian';
		const filePath = `${obsidianDir}/${fileName}`;
		await execAsync(`mkdir -p "${obsidianDir}"`);

		// Create temp file
		tempFile = `/tmp/fabric-${Date.now()}.txt`;

		// Write formatted content to temp file
		await execAsync(`echo ${escapedFormattedContent} > "${tempFile}"`);

		// Copy from temp file to final location (safer than direct write)
		await execAsync(`cp "${tempFile}" "${filePath}"`);

		// Return success response with file details
		return json({
			success: true,
			fileName,
			filePath,
			message: `Successfully saved to ${fileName}`
		});
	} catch (error) {
		console.error('\n=== Error ===');
		console.error('Type:', error?.constructor?.name);
		console.error('Message:', error instanceof Error ? error.message : String(error));
		console.error('Stack:', error instanceof Error ? error.stack : 'No stack trace');

		return json(
			{
				error: error instanceof Error ? error.message : 'Failed to process request',
				details: error instanceof Error ? error.stack : undefined
			},
			{ status: 500 }
		);
	} finally {
		// Clean up temp file if it exists
		if (tempFile) {
			try {
				await execAsync(`rm -f "${tempFile}"`);
			} catch (cleanupError) {
				console.error('Failed to clean up temp file:', cleanupError);
			}
		}
	}
};
