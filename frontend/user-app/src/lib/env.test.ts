import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

describe('public environment defaults', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_API_BASE_URL', '');
    vi.stubEnv('VITE_APP_ENV', '');
    vi.stubEnv('VITE_GRPC_WEB_BASE_URL', '');
    vi.stubEnv('VITE_GRPC_WEB_TIMEOUT_MS', '');
    vi.stubEnv('VITE_PAYMENT_PROVIDERS', '');
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('uses the API Gateway gRPC-Web listener by default', async () => {
    const { env } = await import('./env');

    expect(env.apiBaseUrl).toBe('http://localhost:8080');
    expect(env.grpcWebBaseUrl).toBe('http://localhost:8099');
    expect(env.grpcWebTimeoutMs).toBe(5000);
  });
});
