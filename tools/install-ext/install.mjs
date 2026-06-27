/**
 * Install unpacked extensions into a Chrome profile via CDP Extensions.loadUnpacked.
 * Spawns Chrome with --remote-debugging-pipe and cwd set to the Chrome folder
 * (required for bundled portable Chrome; chrome-launcher does not set cwd).
 *
 * Usage: node install.mjs <config.json>
 */
import fs from "node:fs";
import path from "node:path";
import { spawn, spawnSync } from "node:child_process";
import { Launcher } from "chrome-launcher";

const configPath = process.argv[2];
if (!configPath) {
  console.error("usage: node install.mjs <config.json>");
  process.exit(1);
}

/** @type {{ chrome: string, chromeDir: string, logDir: string, chromeArgs: string[], extensions: { id: string, name: string, path: string }[] }} */
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));

if (!config.extensions?.length) {
  process.exit(0);
}

const logDir = config.logDir || config.chromeDir;
fs.mkdirSync(logDir, { recursive: true });
const outLog = path.join(logDir, "chrome-install-out.log");
const errLog = path.join(logDir, "chrome-install-err.log");
const outFd = fs.openSync(outLog, "a");
const errFd = fs.openSync(errLog, "a");

const baseFlags = Launcher.defaultFlags().filter((f) => f !== "--disable-extensions");
const chromeFlags = [
  ...baseFlags,
  ...config.chromeArgs,
  "--headless=new",
  "--disable-gpu",
  "--remote-debugging-pipe",
  "--enable-unsafe-extension-debugging",
  "about:blank",
];

const proc = spawn(config.chrome, chromeFlags, {
  cwd: config.chromeDir,
  stdio: ["ignore", outFd, errFd, "pipe", "pipe"],
  env: process.env,
});

const pipes = {
  incoming: proc.stdio[4],
  outgoing: proc.stdio[3],
};

/** @type {Map<number, { resolve: (v: unknown) => void, reject: (e: Error) => void }>} */
const pending = new Map();
let buffer = "";
let chromeExited = false;
let exitCode = null;

proc.on("exit", (code) => {
  chromeExited = true;
  exitCode = code;
  for (const [, p] of pending) {
    p.reject(chromeExitError());
  }
  pending.clear();
});

pipes.incoming.on("data", (chunk) => {
  buffer += chunk;
  let end;
  while ((end = buffer.indexOf("\0")) !== -1) {
    const raw = buffer.slice(0, end);
    buffer = buffer.slice(end + 1);
    let msg;
    try {
      msg = JSON.parse(raw);
    } catch {
      continue;
    }
    if (!msg.id || !pending.has(msg.id)) {
      continue;
    }
    const p = pending.get(msg.id);
    pending.delete(msg.id);
    if (msg.error) {
      p.reject(new Error(msg.error.message));
    } else {
      p.resolve(msg.result);
    }
  }
});

function chromeExitError() {
  let detail = "";
  try {
    detail = fs.readFileSync(errLog, "utf8").trim();
  } catch {
    // ignore
  }
  const tail = detail ? `\n${detail.slice(-4000)}` : "";
  return new Error(`Chrome exited (code ${exitCode ?? "unknown"})${tail}`);
}

function assertChromeRunning() {
  if (chromeExited) {
    throw chromeExitError();
  }
}

/**
 * @param {string} method
 * @param {Record<string, unknown>} [params]
 */
function cdpCall(method, params = {}) {
  assertChromeRunning();
  return new Promise((resolve, reject) => {
    const id = Math.floor(Math.random() * 1e9);
    pending.set(id, { resolve, reject });
    pipes.outgoing.write(JSON.stringify({ id, method, params }) + "\0");
    setTimeout(() => {
      if (!pending.has(id)) {
        return;
      }
      pending.delete(id);
      reject(new Error(`cdp ${method}: timed out`));
    }, 90_000);
  });
}

async function killChrome() {
  if (chromeExited) {
    return;
  }
  if (process.platform === "win32") {
    spawnSync(`taskkill /pid ${proc.pid} /T /F`, { shell: true, stdio: "ignore" });
  } else {
    proc.kill("SIGKILL");
  }
  await new Promise((r) => proc.on("exit", r));
}

try {
  await new Promise((r) => setTimeout(r, 1500));
  assertChromeRunning();

  await cdpCall("Browser.getVersion");

  for (const ext of config.extensions) {
    console.log(`Installing extension ${ext.name} (${ext.id}) via CDP`);
    const result = await cdpCall("Extensions.loadUnpacked", { path: ext.path });
    if (result.id !== ext.id) {
      throw new Error(
        `install ${ext.name}: expected id ${ext.id}, got ${result.id}`,
      );
    }
    console.log(`Installed ${ext.name} (${result.id})`);
  }

  await new Promise((r) => setTimeout(r, 3000));
} catch (err) {
  console.error(err.message);
  throw err;
} finally {
  await killChrome();
  fs.closeSync(outFd);
  fs.closeSync(errFd);
}
