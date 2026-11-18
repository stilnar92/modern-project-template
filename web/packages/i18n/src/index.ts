// The texts of all frontends (frontend standard 28, ADR 0012). Paraglide
// compiles messages/*.json into typed functions in src/paraglide: an app
// imports only the messages it uses. Run "mise run gen" after a change of the
// messages. Do not edit src/paraglide by hand.
export { m } from "./paraglide/messages.js";
export { getLocale, locales, baseLocale } from "./paraglide/runtime.js";
