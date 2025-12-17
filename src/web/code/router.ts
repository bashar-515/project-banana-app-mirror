import type { View } from "./views/view";

let currentView: View<unknown> | null = null;

export function switchTo<Type>(
  view: View<Type>,
  state: Type,
) {
  currentView?.tearDown();
  currentView = view;
  currentView.init(state);
}
