import { expect, mock, test } from "bun:test";

import { createClient } from "../src/client/client";
import { getHealth, postV1Scans } from "../src/client";

function lastRequest(mockFetch: ReturnType<typeof mock>) {
  const calls = mockFetch.mock.calls as unknown as [
    input: RequestInfo | URL,
    init?: RequestInit,
  ][];
  const [input, init] = calls[calls.length - 1]!;
  if (input instanceof Request) {
    return input;
  }
  return new Request(input, init);
}

test("getHealth issues GET to {baseUrl}/health with Authorization", async () => {
  const fetchMock = mock((_input: RequestInfo | URL, _init?: RequestInit) =>
    Promise.resolve(
      new Response(
        JSON.stringify({
          status: "healthy",
          service: "safeurl-api",
          checks: {
            database: { status: "healthy" },
            queue: { status: "healthy" },
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      ),
    ),
  );

  const client = createClient({
    baseUrl: "https://api.example.com",
    headers: {
      Authorization: "Bearer sk_live_unit",
    },
    fetch: fetchMock as typeof fetch,
  });

  const result = await getHealth({ client });
  expect(result.response.status).toBe(200);

  const req = lastRequest(fetchMock);
  expect(req.method).toBe("GET");
  expect(req.url).toBe("https://api.example.com/health");
  expect(req.headers.get("Authorization")).toBe("Bearer sk_live_unit");
});

test("createClient normalizes baseUrl trailing slash before requests", async () => {
  const fetchMock = mock((_input: RequestInfo | URL, _init?: RequestInit) =>
    Promise.resolve(
      new Response(
        JSON.stringify({
          status: "healthy",
          service: "safeurl-api",
          checks: {
            database: { status: "healthy" },
            queue: { status: "healthy" },
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      ),
    ),
  );

  const client = createClient({
    baseUrl: "https://api.example.com/",
    fetch: fetchMock as typeof fetch,
  });

  await getHealth({ client });
  const req = lastRequest(fetchMock);
  expect(req.url).toBe("https://api.example.com/health");
});

test("postV1Scans sends JSON body and Content-Type application/json", async () => {
  const fetchMock = mock((_input: RequestInfo | URL, _init?: RequestInit) =>
    Promise.resolve(
      new Response(
        JSON.stringify({
          id: "00000000-0000-0000-0000-000000000001",
          state: "QUEUED",
        }),
        {
          status: 201,
          headers: { "Content-Type": "application/json" },
        },
      ),
    ),
  );

  const client = createClient({
    baseUrl: "https://api.example.com",
    headers: { Authorization: "Bearer sk_live_unit" },
    fetch: fetchMock as typeof fetch,
  });

  const result = await postV1Scans({
    client,
    body: { url: "https://example.com" },
  });
  expect(result.response.status).toBe(201);

  const req = lastRequest(fetchMock);
  expect(req.method).toBe("POST");
  expect(req.url).toBe("https://api.example.com/v1/scans/");
  expect(req.headers.get("Content-Type")).toBe("application/json");
  expect(await req.text()).toBe(JSON.stringify({ url: "https://example.com" }));
});
