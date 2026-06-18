import type { Message } from '@bufbuild/protobuf';
import {
  fileDesc,
  messageDesc,
  serviceDesc,
  type GenMessage,
  type GenService,
} from '@bufbuild/protobuf/codegenv2';

/**
 * Generated-style descriptor for ecommerce/recommendation/v1/recommendation.proto.
 * Replace this file with Buf output when the canonical proto module lands.
 */
export const file_ecommerce_recommendation_v1_recommendation = fileDesc(
  'CjBlY29tbWVyY2UvcmVjb21tZW5kYXRpb24vdjEvcmVjb21tZW5kYXRpb24ucHJvdG8SG2Vjb21tZXJjZS5yZWNvbW1lbmRhdGlvbi52MSKNAQoVUmVjb21tZW5kYXRpb25Db250ZXh0EhsKCXBhZ2VfdHlwZRgBIAEoCVIIcGFnZVR5cGUSHQoKcHJvZHVjdF9pZBgCIAEoCVIJcHJvZHVjdElkEh8KC2NhdGVnb3J5X2lkGAMgASgJUgpjYXRlZ29yeUlkEhcKB3VzZXJfaWQYBCABKAlSBnVzZXJJZCJ/ChlHZXRSZWNvbW1lbmRhdGlvbnNSZXF1ZXN0EkwKB2NvbnRleHQYASABKAsyMi5lY29tbWVyY2UucmVjb21tZW5kYXRpb24udjEuUmVjb21tZW5kYXRpb25Db250ZXh0Ugdjb250ZXh0EhQKBWxpbWl0GAIgASgFUgVsaW1pdCKjAQoSUmVjb21tZW5kYXRpb25JdGVtEh0KCnByb2R1Y3RfaWQYASABKAlSCXByb2R1Y3RJZBIUCgV0aXRsZRgCIAEoCVIFdGl0bGUSGwoJaW1hZ2VfdXJsGAMgASgJUghpbWFnZVVybBIWCgZyZWFzb24YBCABKAlSBnJlYXNvbhIjCg1wcmljZV9kaXNwbGF5GAUgASgJUgxwcmljZURpc3BsYXkiggEKGkdldFJlY29tbWVuZGF0aW9uc1Jlc3BvbnNlEkUKBWl0ZW1zGAEgAygLMi8uZWNvbW1lcmNlLnJlY29tbWVuZGF0aW9uLnYxLlJlY29tbWVuZGF0aW9uSXRlbVIFaXRlbXMSHQoKcmVxdWVzdF9pZBgCIAEoCVIJcmVxdWVzdElkMp8BChVSZWNvbW1lbmRhdGlvblNlcnZpY2UShQEKEkdldFJlY29tbWVuZGF0aW9ucxI2LmVjb21tZXJjZS5yZWNvbW1lbmRhdGlvbi52MS5HZXRSZWNvbW1lbmRhdGlvbnNSZXF1ZXN0GjcuZWNvbW1lcmNlLnJlY29tbWVuZGF0aW9uLnYxLkdldFJlY29tbWVuZGF0aW9uc1Jlc3BvbnNlYgZwcm90bzM=',
);

export type RecommendationContext =
  Message<'ecommerce.recommendation.v1.RecommendationContext'> & {
    pageType: string;
    productId: string;
    categoryId: string;
    userId: string;
  };

export type GetRecommendationsRequest =
  Message<'ecommerce.recommendation.v1.GetRecommendationsRequest'> & {
    context?: RecommendationContext;
    limit: number;
  };

export type RecommendationItem =
  Message<'ecommerce.recommendation.v1.RecommendationItem'> & {
    productId: string;
    title: string;
    imageUrl: string;
    reason: string;
    priceDisplay: string;
  };

export type GetRecommendationsResponse =
  Message<'ecommerce.recommendation.v1.GetRecommendationsResponse'> & {
    items: RecommendationItem[];
    requestId: string;
  };

export const RecommendationContextSchema: GenMessage<RecommendationContext> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 0);

export const GetRecommendationsRequestSchema: GenMessage<GetRecommendationsRequest> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 1);

export const RecommendationItemSchema: GenMessage<RecommendationItem> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 2);

export const GetRecommendationsResponseSchema: GenMessage<GetRecommendationsResponse> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 3);

export const RecommendationService: GenService<{
  getRecommendations: {
    methodKind: 'unary';
    input: typeof GetRecommendationsRequestSchema;
    output: typeof GetRecommendationsResponseSchema;
  };
}> = serviceDesc(file_ecommerce_recommendation_v1_recommendation, 0);
