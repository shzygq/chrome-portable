/**
 * Install unpacked extensions into a Chrome profile via CDP Extensions.loadUnpacked.
 * Uses chrome-launcher --remote-debugging-pipe (works on Windows; Go ExtraFiles does not).
 *
 * Usage: node install.mjs <config.json>
 */
import fs from "node:fs";
import { launch } from "chrome-launcher";

const configPath = process.argv[2];
if (!configPath) {
  console.error("usage: node install.mjs <config.json>");
  process.exit(1);
}

/** @type {{ chrome: string, chromeArgs: string[], extensions: { id: string, name: string, path: string }[] }} */
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));

if (!config.extensions?.length) {
  process.exit(0);
}

const chrome = await launch({
  chromePath: config.chrome,
  chromeFlags: config.chromeArgs,
  ignoreDefaultFlags: true,
  startingUrl: "about:blank",
});

const pipes = chrome.remoteDebuggingPipes;
if (!pipes) {
  await chrome.kill();
  throw new Error("remote debugging pipes unavailable");
}

/**
 * @param {string} method
 * @param {Record<string, unknown>} [params]
 */
function cdpCall(method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = Math.floor(Math.random() * 1e9);
    let buffer = "";

    const onError = (err) => {
      cleanup();
      reject(err);
    };
    const onClose = () => {
      cleanup();
      reject(new Error(`pipe closed before ${method} response`));
    };
    const onData = (chunk) => {
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
        if (msg.id !== id) {
          continue;
        }
        cleanup();
        if (msg.error) {
          reject(new Error(`${method}: ${msg.error.message}`));
          return;
        }
        resolve(msg.result);
        return;
      }
    };

    function cleanup() {
      pipes.incoming.off("error", onError);
      pipes.incoming.off("close", onClose);
      pipes.incoming.off("data", onData);
    }

    pipes.incoming.on("error", onError);
    pipes.incoming.on("close", onClose);
    pipes.incoming.on("data", onData);
    pipes.outgoing.write(JSON.stringify({ id, method, params }) + "\0");
  });
}

try {
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
} finally {
  await chrome.kill();
}
