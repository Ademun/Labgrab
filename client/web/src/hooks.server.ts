import { env } from '$env/dynamic/public';

export async function handleFetch({ event, request, fetch }) {
	if (request.url.startsWith(env.PUBLIC_API_BASE_URL!)) {
		request.headers.set('Cookie', event.request.headers.get('Cookie') ?? '');
	}

	return fetch(request);
}
