import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

import { resetAuthStoreForTest } from "../src/stores/auth-store";

afterEach(() => {
  cleanup();
  resetAuthStoreForTest();
});
