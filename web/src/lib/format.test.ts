import { describe, expect, it } from "vitest";
import { formatDurationMs, serviceTarget } from "./format";

describe("formatDurationMs", () => {
  it("formats milliseconds under a second", () => {
    expect(formatDurationMs(250)).toBe("250ms");
  });

  it("formats seconds", () => {
    expect(formatDurationMs(1500)).toBe("1.50s");
  });
});

describe("serviceTarget", () => {
  it("uses the URL for HTTP services", () => {
    expect(serviceTarget({ type: "http", url: "https://x", hostname: "", port: 0 })).toBe("https://x");
  });

  it("joins host and port for TCP services", () => {
    expect(serviceTarget({ type: "tcp", url: "", hostname: "db.local", port: 5432 })).toBe("db.local:5432");
  });

  it("uses the hostname for DNS services", () => {
    expect(serviceTarget({ type: "dns", url: "", hostname: "example.com", port: 0 })).toBe("example.com");
  });
});
