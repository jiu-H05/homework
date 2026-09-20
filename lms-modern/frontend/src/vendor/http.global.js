(() => {
  // node_modules/@tauri-apps/api/external/tslib/tslib.es6.js
  function __classPrivateFieldGet(receiver, state, kind, f) {
    if (kind === "a" && !f) throw new TypeError("Private accessor was defined without a getter");
    if (typeof state === "function" ? receiver !== state || !f : !state.has(receiver)) throw new TypeError("Cannot read private member from an object whose class did not declare it");
    return kind === "m" ? f : kind === "a" ? f.call(receiver) : f ? f.value : state.get(receiver);
  }
  function __classPrivateFieldSet(receiver, state, value, kind, f) {
    if (kind === "m") throw new TypeError("Private method is not writable");
    if (kind === "a" && !f) throw new TypeError("Private accessor was defined without a setter");
    if (typeof state === "function" ? receiver !== state || !f : !state.has(receiver)) throw new TypeError("Cannot write private member to an object whose class did not declare it");
    return kind === "a" ? f.call(receiver, value) : f ? f.value = value : state.set(receiver, value), value;
  }

  // node_modules/@tauri-apps/api/core.js
  var _Channel_onmessage;
  var _Channel_nextMessageIndex;
  var _Channel_pendingMessages;
  var _Channel_messageEndIndex;
  var _Resource_rid;
  var SERIALIZE_TO_IPC_FN = "__TAURI_TO_IPC_KEY__";
  function transformCallback(callback, once = false) {
    return window.__TAURI_INTERNALS__.transformCallback(callback, once);
  }
  var Channel = class {
    constructor(onmessage) {
      _Channel_onmessage.set(this, void 0);
      _Channel_nextMessageIndex.set(this, 0);
      _Channel_pendingMessages.set(this, []);
      _Channel_messageEndIndex.set(this, void 0);
      __classPrivateFieldSet(this, _Channel_onmessage, onmessage || (() => {
      }), "f");
      this.id = transformCallback((rawMessage) => {
        const index = rawMessage.index;
        if ("end" in rawMessage) {
          if (index == __classPrivateFieldGet(this, _Channel_nextMessageIndex, "f")) {
            this.cleanupCallback();
          } else {
            __classPrivateFieldSet(this, _Channel_messageEndIndex, index, "f");
          }
          return;
        }
        const message = rawMessage.message;
        if (index == __classPrivateFieldGet(this, _Channel_nextMessageIndex, "f")) {
          __classPrivateFieldGet(this, _Channel_onmessage, "f").call(this, message);
          __classPrivateFieldSet(this, _Channel_nextMessageIndex, __classPrivateFieldGet(this, _Channel_nextMessageIndex, "f") + 1, "f");
          while (__classPrivateFieldGet(this, _Channel_nextMessageIndex, "f") in __classPrivateFieldGet(this, _Channel_pendingMessages, "f")) {
            const message2 = __classPrivateFieldGet(this, _Channel_pendingMessages, "f")[__classPrivateFieldGet(this, _Channel_nextMessageIndex, "f")];
            __classPrivateFieldGet(this, _Channel_onmessage, "f").call(this, message2);
            delete __classPrivateFieldGet(this, _Channel_pendingMessages, "f")[__classPrivateFieldGet(this, _Channel_nextMessageIndex, "f")];
            __classPrivateFieldSet(this, _Channel_nextMessageIndex, __classPrivateFieldGet(this, _Channel_nextMessageIndex, "f") + 1, "f");
          }
          if (__classPrivateFieldGet(this, _Channel_nextMessageIndex, "f") === __classPrivateFieldGet(this, _Channel_messageEndIndex, "f")) {
            this.cleanupCallback();
          }
        } else {
          __classPrivateFieldGet(this, _Channel_pendingMessages, "f")[index] = message;
        }
      });
    }
    cleanupCallback() {
      window.__TAURI_INTERNALS__.unregisterCallback(this.id);
    }
    set onmessage(handler) {
      __classPrivateFieldSet(this, _Channel_onmessage, handler, "f");
    }
    get onmessage() {
      return __classPrivateFieldGet(this, _Channel_onmessage, "f");
    }
    [(_Channel_onmessage = /* @__PURE__ */ new WeakMap(), _Channel_nextMessageIndex = /* @__PURE__ */ new WeakMap(), _Channel_pendingMessages = /* @__PURE__ */ new WeakMap(), _Channel_messageEndIndex = /* @__PURE__ */ new WeakMap(), SERIALIZE_TO_IPC_FN)]() {
      return `__CHANNEL__:${this.id}`;
    }
    toJSON() {
      return this[SERIALIZE_TO_IPC_FN]();
    }
  };
  async function invoke(cmd, args = {}, options) {
    return window.__TAURI_INTERNALS__.invoke(cmd, args, options);
  }
  _Resource_rid = /* @__PURE__ */ new WeakMap();

  // node_modules/@tauri-apps/plugin-http/dist-js/index.js
  var ERROR_REQUEST_CANCELLED = "Request cancelled";
  async function fetch(input, init) {
    const signal = init?.signal;
    if (signal?.aborted) {
      throw new Error(ERROR_REQUEST_CANCELLED);
    }
    const maxRedirections = init?.maxRedirections;
    const connectTimeout = init?.connectTimeout;
    const proxy = init?.proxy;
    const danger = init?.danger;
    if (init) {
      delete init.maxRedirections;
      delete init.connectTimeout;
      delete init.proxy;
      delete init.danger;
    }
    const headers = init?.headers ? init.headers instanceof Headers ? init.headers : new Headers(init.headers) : new Headers();
    const req = new Request(input, init);
    const buffer = await req.arrayBuffer();
    const data = buffer.byteLength !== 0 ? Array.from(new Uint8Array(buffer)) : null;
    for (const [key, value] of req.headers) {
      if (!headers.get(key)) {
        headers.set(key, value);
      }
    }
    const headersArray = headers instanceof Headers ? Array.from(headers.entries()) : Array.isArray(headers) ? headers : Object.entries(headers);
    const mappedHeaders = headersArray.map(([name, val]) => [
      name,
      typeof val === "string" ? val : val.toString()
    ]);
    if (signal?.aborted) {
      throw new Error(ERROR_REQUEST_CANCELLED);
    }
    const rid = await invoke("plugin:http|fetch", {
      clientConfig: {
        method: req.method,
        url: req.url,
        headers: mappedHeaders,
        data,
        maxRedirections,
        connectTimeout,
        proxy,
        danger
      }
    });
    const abort = () => invoke("plugin:http|fetch_cancel", { rid }).catch(() => {
    });
    if (signal?.aborted) {
      void abort();
      throw new Error(ERROR_REQUEST_CANCELLED);
    }
    signal?.addEventListener("abort", () => void abort());
    const { status, statusText, url, headers: responseHeaders, rid: responseRid } = await invoke("plugin:http|fetch_send", {
      rid
    });
    let bodyDropped = false;
    const dropBody = () => {
      if (bodyDropped)
        return Promise.resolve();
      bodyDropped = true;
      return invoke("plugin:http|fetch_cancel_body", { rid: responseRid }).catch(() => {
      });
    };
    const readChunk = async (controller) => {
      let data2;
      try {
        data2 = await invoke("plugin:http|fetch_read_body", {
          rid: responseRid
        });
      } catch (e) {
        controller.error(e);
        void dropBody();
        return;
      }
      const dataUint8 = new Uint8Array(data2);
      const lastByte = dataUint8[dataUint8.byteLength - 1];
      const actualData = dataUint8.slice(0, dataUint8.byteLength - 1);
      if (lastByte === 1) {
        controller.close();
        return;
      }
      controller.enqueue(actualData);
    };
    const body = [101, 103, 204, 205, 304].includes(status) ? null : new ReadableStream({
      start: (controller) => {
        signal?.addEventListener("abort", () => {
          controller.error(ERROR_REQUEST_CANCELLED);
          void dropBody();
        });
      },
      pull: (controller) => readChunk(controller),
      cancel: () => {
        void dropBody();
      }
    });
    const res = new Response(body, {
      status,
      statusText
    });
    Object.defineProperty(res, "url", { value: url, writable: false });
    Object.defineProperty(res, "headers", {
      value: new Headers(responseHeaders),
      writable: false
    });
    const originalClone = res.clone.bind(res);
    Object.defineProperty(res, "clone", {
      value: () => {
        const cloned = originalClone();
        Object.defineProperty(cloned, "url", { value: url, writable: false });
        Object.defineProperty(cloned, "headers", {
          value: new Headers(responseHeaders),
          writable: false
        });
        return cloned;
      }
    });
    return res;
  }

  // src/vendor/entry.js
  window.__TAURI__ = window.__TAURI__ || {};
  window.__TAURI__.http = window.__TAURI__.http || {};
  window.__TAURI__.http.fetch = fetch;
})();
