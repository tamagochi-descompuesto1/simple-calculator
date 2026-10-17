import { afterEach, describe, expect, it, vi } from "vitest"
import { ApiError, calculate } from "./apiClient"

// Builds a response like the one fetch gives back for a JSON API.
function jsonResponse(body: unknown, status: number = 200): Response {
    return new Response(JSON.stringify(body), {
        status,
        headers: { 'Content-Type': 'application/json' },
    })
}

// Replaces the global fetch with a fake that answers with createResponse().
// It takes a function and not a Response because a body can be read only once:
// every call needs a fresh response.
function mockFetch(createResponse: () => Response) {
    const fetchMock = vi.fn(async () => createResponse())
    vi.stubGlobal('fetch', fetchMock)
    return fetchMock
}

// Put the real fetch back so one test cannot affect the next.
afterEach(() => {
    vi.unstubAllGlobals()
})

describe('calculate', () => {
    it('sends both operands and returns the result', async () => {
        const fetchMock = mockFetch(() => jsonResponse({ result: 9 }))
        
        const result = await calculate('add', 6, 3)

        expect(result).toBe(9)
        expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/add$/), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ a: 6, b: 3 }),
        })
    })

    it('sends one operand and returns the result', async () => {
        const fetchMock = mockFetch(() => jsonResponse({ result: 4 }))

        const result = await calculate('sqrt', 16)

        expect(result).toBe(4)
        expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/sqrt$/), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json'},
            body: JSON.stringify({ a: 16 }),
        })
    })

    it('sends incorrect operands and returns error', async () => {
        mockFetch(() => jsonResponse({ error: 'division by zero' }, 422))

        const error = await calculate('divide', 5, 0).catch((e: unknown) => e)

        expect(error).toBeInstanceOf(ApiError)
        expect(error).toMatchObject({ message: 'division by zero', status: 422 })
    })

    it('sends incorrect request and returns not JSON error', async () => {
        mockFetch(() => new Response('404 page not found', { status: 404 }))

        const error = await calculate('sqrt', 16).catch((e: unknown) => e)

        expect(error).toBeInstanceOf(ApiError)
        expect(error).toMatchObject({ message: 'Request failed (404)', status: 404})
    })

    it('rejects a success response whose body is not JSON', async () => {
        mockFetch(() => new Response('<html>not an API</html>', { status: 200 }))

        const promise = calculate('add', 1, 2)

        await expect(promise).rejects.toBeInstanceOf(ApiError)
        await expect(promise).rejects.toMatchObject({
            message: 'Unexpected response from the server',
            status: 200,
        })
    })


    it('returns error on no connection', async () => {
        // fetch rejects (instead of returning a response) when it cannot reach the server.
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
        
        const promise = calculate('add', 1, 2)

        await expect(promise).rejects.toBeInstanceOf(ApiError)
        await expect(promise).rejects.toMatchObject({ message: 'Cannot reach the server' })
        await expect(promise).rejects.toHaveProperty('status', undefined)

    })

    // Each row is one test; the first column only names it. The server answered
    // 200 but the body is not {"result": <number>}.
    it.each([
        ['an empty object', {}],
        ['a result that is not a number', { result: '9'}],
        ['null', null],
    ])('rejects a success response with %s', async (_label, body) =>{
        mockFetch(() => jsonResponse(body))
        
        const promise = calculate('add', 1, 2)

        await expect(promise).rejects.toBeInstanceOf(ApiError)
        await expect(promise).rejects.toMatchObject({ message: 'Unexpected response from the server' })
    })
})