import type { Message } from '@bufbuild/protobuf';
import {
  fileDesc,
  messageDesc,
  serviceDesc,
  type GenMessage,
  type GenService,
} from '@bufbuild/protobuf/codegenv2';

export type SessionEventMetadata = Record<string, string>;

/**
 * Generated-style descriptor for ecommerce/session/v1/session.proto.
 * Replace this file with Buf output when the canonical proto module lands.
 */
export const file_ecommerce_session_v1_session = fileDesc(
  'CiJlY29tbWVyY2Uvc2Vzc2lvbi92MS9zZXNzaW9uLnByb3RvEhRlY29tbWVyY2Uuc2Vzc2lvbi52MSL6AgoSSW5nZXN0RXZlbnRSZXF1ZXN0Eh0KCmV2ZW50X25hbWUYASABKAlSCWV2ZW50TmFtZRIZCghwYWdlX3VybBgCIAEoCVIHcGFnZVVybBIdCgpwcm9kdWN0X2lkGAMgASgJUglwcm9kdWN0SWQSFwoHY2FydF9pZBgEIAEoCVIGY2FydElkElIKCG1ldGFkYXRhGAUgAygLMjYuZWNvbW1lcmNlLnNlc3Npb24udjEuSW5nZXN0RXZlbnRSZXF1ZXN0Lk1ldGFkYXRhRW50cnlSCG1ldGFkYXRhEh8KC29jY3VycmVkX2F0GAYgASgJUgpvY2N1cnJlZEF0EiEKDGFub255bW91c19pZBgHIAEoCVILYW5vbnltb3VzSWQSHQoKc2Vzc2lvbl9pZBgIIAEoCVIJc2Vzc2lvbklkGjsKDU1ldGFkYXRhRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4ASJMChNJbmdlc3RFdmVudFJlc3BvbnNlEhkKCGV2ZW50X2lkGAEgASgJUgdldmVudElkEhoKCGFjY2VwdGVkGAIgASgIUghhY2NlcHRlZDJ0Cg5TZXNzaW9uU2VydmljZRJiCgtJbmdlc3RFdmVudBIoLmVjb21tZXJjZS5zZXNzaW9uLnYxLkluZ2VzdEV2ZW50UmVxdWVzdBopLmVjb21tZXJjZS5zZXNzaW9uLnYxLkluZ2VzdEV2ZW50UmVzcG9uc2ViBnByb3RvMw==',
);

export type IngestEventRequest = Message<'ecommerce.session.v1.IngestEventRequest'> & {
  eventName: string;
  pageUrl: string;
  productId: string;
  cartId: string;
  metadata: SessionEventMetadata;
  occurredAt: string;
  anonymousId: string;
  sessionId: string;
};

export type IngestEventResponse =
  Message<'ecommerce.session.v1.IngestEventResponse'> & {
    eventId: string;
    accepted: boolean;
  };

export const IngestEventRequestSchema: GenMessage<IngestEventRequest> =
  messageDesc(file_ecommerce_session_v1_session, 0);

export const IngestEventResponseSchema: GenMessage<IngestEventResponse> =
  messageDesc(file_ecommerce_session_v1_session, 1);

export const SessionService: GenService<{
  ingestEvent: {
    methodKind: 'unary';
    input: typeof IngestEventRequestSchema;
    output: typeof IngestEventResponseSchema;
  };
}> = serviceDesc(file_ecommerce_session_v1_session, 0);
