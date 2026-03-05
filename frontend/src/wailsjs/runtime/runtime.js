// Wails runtime bridge — auto-overwritten by `wails dev` / `wails build`

export function EventsOn(eventName, callback) {
  return window['runtime']['EventsOn'](eventName, callback);
}

export function EventsOff(...eventNames) {
  return window['runtime']['EventsOff'](...eventNames);
}

export function EventsOnce(eventName, callback) {
  return window['runtime']['EventsOnce'](eventName, callback);
}

export function EventsEmit(eventName, ...data) {
  return window['runtime']['EventsEmit'](eventName, ...data);
}

export function WindowMinimise() {
  return window['runtime']['WindowMinimise']();
}

export function WindowMaximise() {
  return window['runtime']['WindowMaximise']();
}

export function Quit() {
  return window['runtime']['Quit']();
}
