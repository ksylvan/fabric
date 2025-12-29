import { createPipeline, transformers } from 'pdf-to-markdown-core/lib/src';
import { PARSE_SCHEMA } from 'pdf-to-markdown-core/lib/src/PdfParser';

// pdfjs-dist v5+ requires browser APIs at import time, so we use dynamic imports
let pdfjs: typeof import('pdfjs-dist') | null = null;

export class PdfConversionService {
	private async ensureInitialized(): Promise<typeof import('pdfjs-dist')> {
		if (!pdfjs) {
			// Dynamic import to avoid SSR issues with pdfjs-dist v5+
			pdfjs = await import('pdfjs-dist');
			const pdfConfig = (await import('./pdf-config')).default;
			await pdfConfig.initialize();
		}
		return pdfjs;
	}

	async convertToMarkdown(file: File): Promise<string> {
		const pdfjsLib = await this.ensureInitialized();

		const buffer = await file.arrayBuffer();

		const pipeline = createPipeline(pdfjsLib, {
			transformConfig: {
				transformers
			}
		});

		const result = await pipeline.parse(buffer, () => {});

		const transformed = result.transform();

		const markdown = transformed.convert({
			convert: (items) => {
				const text = items
					.map((item) => item.value('str')) // Using 'str' instead of 'text' based on PARSE_SCHEMA
					.filter(Boolean)
					.join('\n');

				return text;
			}
		});

		return markdown;
	}
}
