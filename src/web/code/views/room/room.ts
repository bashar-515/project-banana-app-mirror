import type { Client } from "../../../../../sdk/client";
import { Session } from "../../session/session";
import type { HomeView } from "../home/home";
import type { View } from "../view";

export interface RoomViewParams {
  roomId: string,
  peerId: string,
  caller: boolean,
}

export class RoomView implements View<RoomViewParams> {
  // HTML elements
  private roomDiv: HTMLElement | null = document.getElementById("roomView");
  private roomIdParagraph: HTMLElement | null = document.getElementById("roomIdParagraph");
  private localVideo: HTMLVideoElement | null = document.getElementById("localVideo") as HTMLVideoElement;
  private remoteVideo: HTMLVideoElement | null = document.getElementById("remoteVideo") as HTMLVideoElement;

  private localStream: MediaStream | null = null;
  private remoteStream: MediaStream | null = null;

  private roomId: string | null = null;

  // state
  private session: Session | null = null;

  private copyRoomId = async (): Promise<void> => {
    try {
      if (this.roomId) await navigator.clipboard.writeText(this.roomId);
    } catch (e) {
      console.error(e);
    }
  };

  constructor(homeView: HomeView, client: Client) {
    this.session = new Session(homeView, client)
  }

  public async init(roomViewParams: RoomViewParams): Promise<void> {
    const { roomId, peerId, caller } = roomViewParams

    this.roomId = roomId;

    this.initUi(roomId);

    this.localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
    this.remoteStream = new MediaStream();

    if (this.localVideo) this.localVideo.srcObject = this.localStream;
    if (this.remoteVideo) this.remoteVideo.srcObject = this.remoteStream;

    await this.session?.start(roomId, peerId, caller, this.localStream, this.remoteStream);
  }

  public tearDown(): void {
    if (this.session) this.session.stop();

    this.remoteStream?.getTracks().forEach((track) => { track.stop(); });
    this.localStream?.getTracks().forEach((track) => { track.stop(); });

    if (this.remoteVideo) this.remoteVideo.srcObject = null;
    if (this.localVideo) this.localVideo.srcObject = null;

    this.tearDownUi();

    this.roomId = null;
  }

  private initUi(roomId: string): void {
    if (this.roomIdParagraph) {
      this.roomIdParagraph.textContent = roomId;
      this.roomIdParagraph.style.cursor = "pointer";
      this.roomIdParagraph.title = "click to copy room ID";
      this.roomIdParagraph.addEventListener("click", this.copyRoomId)
    }

    this.roomDiv?.style.setProperty("display", "block");
  }

  private tearDownUi(): void {
    this.roomDiv?.style.setProperty("display", "none");

    if (this.roomIdParagraph) {
      this.roomIdParagraph.removeEventListener("click", this.copyRoomId)
      this.roomIdParagraph.title = "";
      this.roomIdParagraph.style.cursor = "";
      this.roomIdParagraph.textContent = "";
    }
  }
}
