export const MAX_NATIVE_MULTIPART_BYTES: number;
export const MAX_NATIVE_MULTIPART_PARTS: number;
export interface PreparedNativeMultipart {
  readonly body: ArrayBuffer;
  readonly contentType: string;
}
export function prepareNativeMultipart(form: FormData): Promise<PreparedNativeMultipart>;
