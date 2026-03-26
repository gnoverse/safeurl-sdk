import { expect, test } from "bun:test";

import { createClient } from "../src/client/client";
import {
  deleteV1ApiKeysById,
  getHealth,
  getV1ApiKeys,
  getV1Credits,
  getV1ScansById,
  postV1ApiKeys,
  postV1CreditsPurchase,
  postV1Scans,
} from "../src/client";

type CreditsResponse = {
  balance: number;
  userId: string;
  updatedAt: string;
};

type PurchaseCreditsResponse = {
  id: string;
  userId: string;
  amount: number;
  status: string;
  createdAt: string;
  completedAt: string;
  newBalance: number;
};

type ScanResponse = {
  id: string;
  url: string;
  state: "QUEUED" | "FETCHING" | "ANALYZING" | "COMPLETED" | "FAILED" | "TIMED_OUT";
  createdAt: string;
  updatedAt: string;
};

type CreateApiKeyResponse = {
  id: string;
  key: string;
  name: string;
  createdAt: string;
};

type ApiKeyListItem = {
  id: string;
  name: string;
  scopes: string[];
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdAt: string;
  revokedAt: string | null;
};

const baseUrl = process.env["SAFEURL_SDK_TEST_BASE_URL"] ?? "http://localhost:8081";
const serviceSecret =
  process.env["SAFEURL_SDK_TEST_SERVICE_SECRET"] ??
  process.env["SAFEURL_SERVICE_SECRET"];
const smoke = serviceSecret ? test : test.skip;

smoke("bootstraps and uses a real API key", async () => {
  if (!serviceSecret) {
    return;
  }

  const serviceClient = createClient({
    baseUrl,
    headers: {
      Authorization: `Bearer ${serviceSecret}`,
    },
  });

  const health = await getHealth({
    client: serviceClient,
  });

  expect(health.response.status).toBe(200);
  expect(health.data).toMatchObject({
    status: "healthy",
    service: "safeurl-api",
    checks: {
      database: { status: "healthy" },
      queue: { status: "healthy" },
    },
  });

  let createdKey: CreateApiKeyResponse | undefined;
  let revokeKey: CreateApiKeyResponse | undefined;

  try {
    const keyName = `ts-sdk-smoke-${Date.now()}`;
    const createApiKeyResult = await postV1ApiKeys({
      client: serviceClient,
      body: {
        name: keyName,
        scopes: ["scan:read", "scan:write", "credits:read", "credits:write"],
      },
    });
    expect(createApiKeyResult.response.status).toBe(201);
    createdKey = createApiKeyResult.data as CreateApiKeyResponse;
    expect(createdKey.id).toBeTruthy();
    expect(createdKey.name).toBe(keyName);
    expect(createdKey.key).toMatch(/^sk_live_/);

    const apiKeyClient = createClient({
      baseUrl,
      headers: {
        Authorization: `Bearer ${createdKey.key}`,
      },
    });

    const apiKeyListResult = await getV1ApiKeys({
      client: apiKeyClient,
    });
    expect(apiKeyListResult.response.status).toBe(200);
    const apiKeyList = apiKeyListResult.data as ApiKeyListItem[];
    expect(
      apiKeyList.some((key) => key.id === createdKey?.id && key.name === keyName),
    ).toBe(true);

    const purchaseResult = await postV1CreditsPurchase({
      client: serviceClient,
      body: {
        amount: 1,
      },
    });
    expect(purchaseResult.response.status).toBe(201);
    const purchase = purchaseResult.data as PurchaseCreditsResponse;
    expect(purchase.amount).toBe(1);

    const apiKeyCreditsResult = await getV1Credits({
      client: apiKeyClient,
    });
    expect(apiKeyCreditsResult.response.status).toBe(200);
    const apiKeyCredits = apiKeyCreditsResult.data as CreditsResponse;
    expect(apiKeyCredits.balance).toBe(purchase.newBalance);

    const revokeKeyName = `ts-sdk-revoke-${Date.now()}`;
    const createRevokeKeyResult = await postV1ApiKeys({
      client: serviceClient,
      body: {
        name: revokeKeyName,
        scopes: ["scan:read", "credits:read"],
      },
    });
    expect(createRevokeKeyResult.response.status).toBe(201);
    revokeKey = createRevokeKeyResult.data as CreateApiKeyResponse;
    expect(revokeKey.id).toBeTruthy();
    expect(revokeKey.key).toMatch(/^sk_live_/);

    const revokeClient = createClient({
      baseUrl,
      headers: {
        Authorization: `Bearer ${revokeKey.key}`,
      },
    });

    const revokePrecheck = await getV1Credits({
      client: revokeClient,
    });
    expect(revokePrecheck.response.status).toBe(200);

    const revokeResult = await deleteV1ApiKeysById({
      client: apiKeyClient,
      path: { id: revokeKey.id },
    });
    expect(revokeResult.response.status).toBe(200);
    revokeKey = undefined;

    const revokedCreditsResult = await getV1Credits({
      client: revokeClient,
    });
    expect(revokedCreditsResult.response.status).toBe(401);

    const unauthenticatedClient = createClient({ baseUrl });
    const missingAuthResult = await getV1Credits({
      client: unauthenticatedClient,
    });
    expect(missingAuthResult.response.status).toBe(401);

    const invalidAuthClient = createClient({
      baseUrl,
      headers: {
        Authorization: "Bearer sk_live_invalid_api_key",
      },
    });
    const invalidAuthResult = await getV1Credits({
      client: invalidAuthClient,
    });
    expect(invalidAuthResult.response.status).toBe(401);

    const scanUrl = "https://example.com";
    const createScanResult = await postV1Scans({
      client: apiKeyClient,
      body: {
        url: scanUrl,
      },
    });
    expect(createScanResult.response.status).toBe(201);

    const createdScan = createScanResult.data as { id: string; state: "QUEUED" };
    expect(createdScan.id).toBeTruthy();
    expect(createdScan.state).toBe("QUEUED");

    const scanResult = await getV1ScansById({
      client: apiKeyClient,
      path: { id: createdScan.id },
    });
    expect(scanResult.response.status).toBe(200);

    const scan = scanResult.data as ScanResponse;
    expect(scan.id).toBe(createdScan.id);
    expect(scan.url).toBe(scanUrl);
    expect([
      "QUEUED",
      "FETCHING",
      "ANALYZING",
      "COMPLETED",
      "FAILED",
      "TIMED_OUT",
    ]).toContain(scan.state);
    expect(scan.createdAt).toBeDefined();
    expect(scan.updatedAt).toBeDefined();
  } finally {
    if (revokeKey?.id) {
      await deleteV1ApiKeysById({
        client: serviceClient,
        path: { id: revokeKey.id },
      });
    }
    if (createdKey?.id) {
      await deleteV1ApiKeysById({
        client: serviceClient,
        path: { id: createdKey.id },
      });
    }
  }
}, 30_000);
