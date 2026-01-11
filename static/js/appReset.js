const EVT = "app:reset";

export function appReset(reason = "unknown") {
  window.dispatchEvent(new CustomEvent(EVT, { detail: { reason } }));
}

export function onAppReset(fn) {
  window.addEventListener(EVT, fn);
}
