import { browser } from "$app/environment";

// Export the configuration - accepts pdfjs module to avoid top-level import
// This is necessary because pdfjs-dist v5+ uses browser APIs at import time
export default {
	initialize: async () => {
		if (browser) {
			// Dynamic import to avoid SSR issues
			const pdfjs = await import("pdfjs-dist");
			const { GlobalWorkerOptions, version } = pdfjs;

			// Use CDN-hosted worker to avoid bundling third-party minified code in the repo
			const workerSrc = `https://unpkg.com/pdfjs-dist@${version}/build/pdf.worker.min.mjs`;
			GlobalWorkerOptions.workerSrc = workerSrc;
		}
	},
};
