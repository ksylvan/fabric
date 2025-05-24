import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
    const fabUrl = process.env.FABRIC_BASE_URL || 'http://localhost:8080';

    // Pass the FABRIC_BASE_URL to the client by injecting it into the HTML
    // SvelteKit's %sveltekit.head% placeholder is a good place for this.
    // We'll use a script tag to set a global variable.
    const response = await resolve(event, {
        transformPageChunk: ({ html }) =>
        html.replace(
            '%sveltekit.head%',
            `%sveltekit.head%
            <script>
            window.__FABRIC_CONFIG__ = { FABRIC_BASE_URL: '${fabUrl}' };
            </script>`
        )
    });
    return response;
};
