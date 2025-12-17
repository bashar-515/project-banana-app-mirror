import type { View } from "../view";

export class HomeView implements View {
  private homeDiv: HTMLElement | null = document.getElementById("homeView");

  init(): void {
    this.homeDiv?.style.setProperty("display", "block");
  }

  tearDown(): void {
    this.homeDiv?.style.setProperty("display", "none");
  }
}
