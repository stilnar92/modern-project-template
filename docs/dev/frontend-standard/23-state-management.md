# 23 — State Management

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### Subscribing to external state

Source: [React Tricks](https://molefrog.com/notes/react-tricks) (Alexey Taktarov).

- **`useSyncExternalStore` is the way to subscribe to anything outside React** (stores, WebSockets, browser events like `hashchange`): `useState` + `useEffect` subscriptions risk tearing under concurrent rendering and don't work in SSR; `useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)` solves both.
- **Stable `subscribe` function required** — an inline anonymous subscribe causes re-subscription every render; hoist it to module scope or stabilize via a `useState` initializer.
- **Async-event tearing trick** — when a browser event fires asynchronously per listener (e.g. `hashchange` in WebKit), subscribe once and re-dispatch a synchronous manual event so all `useSyncExternalStore` callbacks run in the same task.
