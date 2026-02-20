export async function handleFetch({ event, request, fetch }) {
    if (request.url.startsWith('http://127.0.0.1:8080')) {
        console.log(request.headers)
        request.headers.set('Cookie', event.request.headers.get('Cookie') ?? '');
    }

    return fetch(request);
}