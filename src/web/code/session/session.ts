import type { Client } from "../../../../sdk/client";
import { RTC_CONFIG } from "../config/webrtc";
import { switchTo } from "../router";
import { makeWebSocketUrl } from "../utils"
import type { HomeView } from "../views/home/home";
import { mapType } from "./helpers";

interface DescriptionMessage {
  message: string;
  description: {
    sdp: string;
    type: number
  }
}

interface CandidatesMessage {
  message: string;
  candidates: Candidate[];
}

interface CandidateMessage {
  message: string;
  candidate: Candidate;
}

interface Candidate {
  candidate: string;
  sdpMLineIndex: string;
  sdpMid: string;
  usernameFragment: string;
}

export class Session {
  private homeView: HomeView | null = null;
  private client: Client | null = null;

  private wsConnection: WebSocket | null = null;
  private peerConnection: RTCPeerConnection | null = null;
  private remoteDescriptionSet: Promise<void>;
  private setRemoteDescriptionResolver: (() => void) | null = null;

  constructor(homeView: HomeView, client: Client) {
    this.client = client;
    this.homeView = homeView;
    this.remoteDescriptionSet = new Promise<void>((resolve) => {
      this.setRemoteDescriptionResolver = resolve;
    });
  }

  public async start(roomId: string, peerId: string, caller: boolean, localStream: MediaStream, remoteStream: MediaStream): Promise<void> {
    this.peerConnection = new RTCPeerConnection(RTC_CONFIG);

    this.peerConnection.addEventListener('connectionstatechange', () => {
    if (
      this.peerConnection &&
      (this.peerConnection.connectionState === "disconnected" ||
        this.peerConnection.connectionState === "failed" ||
        this.peerConnection.connectionState === "closed")
    ) {
      this.handlePeerDisconnect();
    }
  });

  this.peerConnection.addEventListener('iceconnectionstatechange', () => {
    if (
      this.peerConnection &&
      (this.peerConnection.iceConnectionState === "disconnected" ||
        this.peerConnection.iceConnectionState === "failed" ||
        this.peerConnection.iceConnectionState === "closed")
    ) {
      this.handlePeerDisconnect();
    }
  });

    localStream.getTracks().forEach((track) => {
      this.peerConnection?.addTrack(track, localStream);
    });

    if (this.peerConnection) {
      this.peerConnection.ontrack = event => {
        event.streams[0].getTracks().forEach((track) => {
          remoteStream.addTrack(track);
        });
      }
    }

    this.peerConnection.onicecandidate = event => {
      if (event.candidate) this.client?.uploadIceCandidate(roomId, peerId, event.candidate);
    }

    if (caller) {
      const sessionDescriptionOffer = await this.peerConnection.createOffer();

      await this.peerConnection.setLocalDescription(sessionDescriptionOffer)
      await this.client?.uploadSessionDescription(roomId, peerId, sessionDescriptionOffer)

      this.wsConnection = new WebSocket(makeWebSocketUrl(roomId, peerId));

      this.wsConnection.addEventListener("message", (event) => {
        try {
          const data = JSON.parse(event.data) as DescriptionMessage;

          if (data.message == "description") {
            this.handleAnswerDescriptionMessage(data)
          } 
        } catch (e) {
          console.error(e)
        }
      });
    } else { 
      this.wsConnection = new WebSocket(makeWebSocketUrl(roomId, peerId));

      this.wsConnection.addEventListener("message", async (event) => {
        try {
          const data = JSON.parse(event.data) as DescriptionMessage;

          if (data.message == "description") {
            this.handleOfferDescriptionMessage(data, roomId, peerId)
          }
        } catch (e) {
          console.error(e)
        }
      });
    }

    this.wsConnection.addEventListener("message", (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.message == "candidates") {
          this.handleCandidatesMessage(data as CandidatesMessage);
        } else if (data.message == "candidate") {
          this.handleCandidateMessage(data as CandidateMessage)
        }
      } catch (e) {
        console.error(e)
      }
    });
  }

  public stop(): void {
    if (this.peerConnection) {
      this.peerConnection.ontrack = null;
      this.peerConnection.onicecandidate = null;
      this.peerConnection.close();
      this.peerConnection = null;
    }

    if (this.wsConnection) {
      this.wsConnection.close();
      this.wsConnection = null;
    }

    this.remoteDescriptionSet = new Promise<void>((resolve) => {
      this.setRemoteDescriptionResolver = resolve;
    });
  }

  private handlePeerDisconnect(): void {
    switchTo(this.homeView!, undefined) 
  }

  private handleAnswerDescriptionMessage(data: DescriptionMessage): void {
    if (!this.peerConnection?.currentRemoteDescription) {
      const sessionDescriptionAnswer = new RTCSessionDescription({
        sdp: data.description.sdp === "" ? undefined : data.description.sdp,
        type: mapType(data.description.type),
      })

      this.peerConnection?.setRemoteDescription(sessionDescriptionAnswer);
      this.setRemoteDescriptionResolver?.();
    }
  }

  private async handleOfferDescriptionMessage(data: DescriptionMessage, roomId: string, peerId: string): Promise<void> {
    const sessionDescriptionOffer = new RTCSessionDescription({
      sdp: data.description.sdp === "" ? undefined : data.description.sdp,
      type: mapType(data.description.type),
    });

    this.peerConnection?.setRemoteDescription(sessionDescriptionOffer);
    this.setRemoteDescriptionResolver?.();

    if (this.peerConnection) {
      const sessionDescriptionAnswer = await this.peerConnection.createAnswer();

      await this.peerConnection.setLocalDescription(sessionDescriptionAnswer);
      await this.client?.uploadSessionDescription(roomId, peerId, sessionDescriptionAnswer);
    }
  }

  private async handleCandidatesMessage(data: CandidatesMessage): Promise<void> {
    await this.remoteDescriptionSet;
    
    for (const candidate of data.candidates) {
      const iceCandidate = new RTCIceCandidate({
        candidate: candidate.candidate === "" ? undefined : candidate.candidate, 
        sdpMLineIndex: candidate.sdpMLineIndex === "" ? undefined : Number(candidate.sdpMLineIndex),
        sdpMid: candidate.sdpMid === "" ? undefined: candidate.sdpMid,
        usernameFragment: candidate.usernameFragment === "" ? undefined: candidate.usernameFragment,
      });

      this.peerConnection?.addIceCandidate(iceCandidate);
    }
  }

  private async handleCandidateMessage(data: CandidateMessage): Promise<void> {
    await this.remoteDescriptionSet;
    
    const iceCandidate = new RTCIceCandidate({
      candidate: data.candidate.candidate === "" ? undefined : data.candidate.candidate, 
      sdpMLineIndex: data.candidate.sdpMLineIndex === "" ? undefined : Number(data.candidate.sdpMLineIndex),
      sdpMid: data.candidate.sdpMid === "" ? undefined: data.candidate.sdpMid,
      usernameFragment: data.candidate.usernameFragment === "" ? undefined: data.candidate.usernameFragment,
    });

    this.peerConnection?.addIceCandidate(iceCandidate);
  }
}
