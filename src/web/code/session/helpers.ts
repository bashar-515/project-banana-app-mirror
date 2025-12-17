export function mapType(num: number): RTCSdpType {
  switch (num) {
    case 1: return "answer";
    case 2: return "offer";
    case 3: return "pranswer";
    case 4: return "rollback";
    default: throw new Error(`unknown SDP type enum: ${num}`);
  }
}
