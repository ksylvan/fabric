// pdfjs-dist v5+ requires browser APIs at import time, so we use dynamic imports
let pdfjs: typeof import("pdfjs-dist") | null = null;

export class PdfConversionService {
	private async ensureInitialized(): Promise<typeof import("pdfjs-dist")> {
		if (!pdfjs) {
			// Dynamic import to avoid SSR issues with pdfjs-dist v5+
			pdfjs = await import("pdfjs-dist");
			const pdfConfig = (await import("./pdf-config")).default;
			console.log("PDF.js version:", pdfjs.version);
			await pdfConfig.initialize();
			console.log("Worker configuration complete");
		}
		return pdfjs;
	}

	async convertToMarkdown(file: File): Promise<string> {
		console.log("Starting PDF conversion:", {
			fileName: file.name,
			fileSize: file.size,
		});

		const pdfjsLib = await this.ensureInitialized();

		const buffer = await file.arrayBuffer();
		console.log("Buffer created:", buffer.byteLength);

		const pdf = await pdfjsLib.getDocument({ data: buffer }).promise;
		console.log("PDF loaded, pages:", pdf.numPages);

		const pages: string[] = [];

		for (let pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
			const page = await pdf.getPage(pageNum);
			const textContent = await page.getTextContent();

			let lastY: number | null = null;
			const lines: string[] = [];
			let currentLine = "";

			for (const item of textContent.items) {
				if (!("str" in item)) continue;
				const textItem = item as { str: string; transform: number[] };
				const y = textItem.transform[5];

				if (lastY !== null && Math.abs(y - lastY) > 2) {
					// New line detected (y position changed)
					if (currentLine.trim()) {
						lines.push(currentLine.trim());
					}
					currentLine = textItem.str;
				} else {
					currentLine += textItem.str;
				}
				lastY = y;
			}
			// Push the last line
			if (currentLine.trim()) {
				lines.push(currentLine.trim());
			}

			if (lines.length > 0) {
				pages.push(lines.join("\n"));
			}
		}

		const markdown = pages.join("\n\n");

		console.log("PDF conversion completed:", {
			resultLength: markdown.length,
			preview: markdown.substring(0, 100),
		});

		return markdown;
	}
}
