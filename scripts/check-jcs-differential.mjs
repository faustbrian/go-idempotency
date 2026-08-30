import { readFileSync } from "node:fs";
import canonicalize from "canonicalize";

const fixture = JSON.parse(
  readFileSync(new URL("../canonical/testdata/rfc8785.json", import.meta.url), "utf8"),
);

for (const test of fixture.cases) {
  const actual = canonicalize(JSON.parse(test.input));
  if (actual !== test.expected) {
    throw new Error(`${test.name}: maintained JCS peer disagrees with pinned result`);
  }
}

process.stdout.write(`maintained JCS peer agreement: ${fixture.cases.length} cases\n`);
