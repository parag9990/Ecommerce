import { beforeEach, describe, expect, it } from "vitest";

import {
  clearAdminSession,
  getAccessToken,
  readAdminSession,
  writeAdminSession
} from "./auth-session";

describe("admin auth session", () => {
  beforeEach(() => clearAdminSession());

  it("stores only an active admin session", () => {
    writeAdminSession({
      accessToken: "token",
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      roles: ["operations_admin"],
      userId: "admin_1"
    });
    expect(readAdminSession()?.userId).toBe("admin_1");
    expect(getAccessToken()).toBe("token");
  });

  it("rejects expired sessions", () => {
    writeAdminSession({
      accessToken: "expired",
      expiresAt: new Date(Date.now() - 1_000).toISOString(),
      roles: ["admin"],
      userId: "admin_1"
    });
    expect(readAdminSession()).toBeNull();
    expect(getAccessToken()).toBeUndefined();
  });
});
