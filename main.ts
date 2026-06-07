import app from "./bin/app.wasm";
import "./bin/wasm_exec.js";

globalThis.tryCatch = (o, fn, args) => {
    try {
        if (fn) {
            return { data: o[fn](...args) };
        }

        return { data: o(...args) };
    } catch (err) {
        if (!(err instanceof Error)) {
            if (err instanceof Object) {
                err = JSON.stringify(err);
            }
            err = new Error(err || "no error message");
        }
        return { error: err };
    }
};

let initiliazed = false;

let go = new Go();
let instance = new WebAssembly.Instance(app, go.importObject);

function init() {
    if (!initiliazed) {
        go.run(instance).finally(() => {
            initiliazed = false;
            instance = new WebAssembly.Instance(app, go.importObject);
        });
        initiliazed = true;
    }

    if (go.exited) {
        go = new Go();
        go.run(instance).finally(() => {
            instance = new WebAssembly.Instance(app, go.importObject);
        });
    }
}

async function fetch(req: Request, env: Env, ctx: ExecutionContext) {
    init()
    return await globalThis.cf.fetch(req, env, ctx);
}

export default {
    fetch,
};
