import { Client } from "../../../sdk/client";
import { switchTo } from "./router";
import { HomeView } from "./views/home/home";
import { RoomView, type RoomViewParams } from "./views/room/room";

export class App {
  private client: Client = new Client("https://api.project-banana.com");

  private homeView: HomeView = new HomeView();
  private roomView: RoomView = new RoomView(this.homeView, this.client);

  async init(): Promise<void> {
    this.wireButtons();

    const roomId = new URLSearchParams(window.location.search).get("roomId");

    if (roomId) {
      const peerId: string = await this.client.joinRoom(roomId);

      const roomViewParams: RoomViewParams = {
        roomId: roomId,
        peerId: peerId,
        caller: false,
      };

      switchTo(this.roomView, roomViewParams);
    } else {
      switchTo(this.homeView, undefined);
    }
  }

  private wireButtons(): void {
    this.wireHomeButtons();
    this.wireRoomButtons();
  }

  private wireHomeButtons(): void {
    const createButton = document.getElementById("createButton");
    const joinButton = document.getElementById("joinButton");

    const roomIdInput = document.getElementById(
      "roomIdInput",
    ) as HTMLInputElement;

    createButton?.addEventListener("click", async () => {
      const roomId: string = await this.client.createRoom();
      const peerId: string = await this.client.joinRoom(roomId);

      const roomViewParams: RoomViewParams = {
        roomId: roomId,
        peerId: peerId,
        caller: true,
      };

      switchTo(this.roomView, roomViewParams);
    });

    joinButton?.addEventListener("click", async () => {
      const roomId: string = roomIdInput?.value.trim();
      const peerId: string = await this.client.joinRoom(roomId);

      if (roomIdInput) roomIdInput.value = "";

      const roomViewParams: RoomViewParams = {
        roomId: roomId,
        peerId: peerId,
        caller: false,
      };

      switchTo(this.roomView, roomViewParams);
    });
  }

  private wireRoomButtons(): void {
    const leaveButton = document.getElementById("leaveButton");

    leaveButton?.addEventListener("click", () => {
      switchTo(this.homeView, undefined);
    });
  }
}
