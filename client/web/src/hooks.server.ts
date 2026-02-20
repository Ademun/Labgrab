import { PUBLIC_API_BASE_URL } from '$env/static/public';

export async function handleFetch({ event, request, fetch }) {
	if (request.url.startsWith(PUBLIC_API_BASE_URL)) {
		request.headers.set('Cookie', event.request.headers.get('Cookie') ?? '');
	}

	return fetch(request);
}
