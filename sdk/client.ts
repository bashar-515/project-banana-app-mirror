import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient, type Client as rpcClient } from "@connectrpc/connect";
import { create } from "@bufbuild/protobuf";
import {
  AppService,
  CreateRoomRequestSchema,
  type CreateRoomRequest,
  type CreateRoomResponse,
  JoinRoomRequestSchema,
  type JoinRoomRequest,
  type JoinRoomResponse,
  UploadIceCandidateRequestSchema,
  type UploadIceCandidateRequest,
  UploadSessionDescriptionRequestSchema,
  type UploadSessionDescriptionRequest,
  Type,
} from "../gen/ts/app/v1/app_pb";

export class Client {
  private client: rpcClient<typeof AppService>;

  constructor(baseUrl: string) {
    const transport = createConnectTransport({ baseUrl: baseUrl });

    this.client = createClient(AppService, transport);
  }

  public async createRoom(): Promise<string> {
    const request: CreateRoomRequest = create(CreateRoomRequestSchema);
    const response: CreateRoomResponse = await this.client.createRoom(request);

    return response.roomId;
  }

  public async joinRoom(roomId: string): Promise<string> {
    const request: JoinRoomRequest = create(JoinRoomRequestSchema, { roomId: roomId });
    const response: JoinRoomResponse = await this.client.joinRoom(request);

    return response.peerId;
  }

  public async uploadIceCandidate(roomId: string, peerId: string, candidate: RTCIceCandidate): Promise<void> {
    const candidateToJson = candidate.toJSON();

    const request: UploadIceCandidateRequest = create(
      UploadIceCandidateRequestSchema,
      {
        roomId: roomId,
        peerId: peerId,
        iceCandidate: {
          candidate: candidateToJson.candidate ?? undefined,
          sdpMLineIndex: candidateToJson.sdpMLineIndex ?? undefined,
          sdpMid: candidateToJson.sdpMid ?? undefined,
          usernameFragment: candidateToJson.usernameFragment ?? undefined,
        },
      },
    );

    await this.client.uploadIceCandidate(request)
  }

  public async uploadSessionDescription(roomId: string, peerId: string, offer: RTCSessionDescriptionInit): Promise<void> {
    const type =
      offer.type === "offer"    ? Type.OFFER :
      offer.type === "answer"   ? Type.ANSWER :
      offer.type === "pranswer" ? Type.PRANSWER :
      offer.type === "rollback" ? Type.ROLLBACK :
                                  Type.UNSPECIFIED;

    const request: UploadSessionDescriptionRequest = create(
      UploadSessionDescriptionRequestSchema,
      {
        roomId: roomId,
        peerId: peerId,
        sessionDescription: { sdp: offer.sdp, type: type },
      },
    );

    await this.client.uploadSessionDescription(request)
  }
}
