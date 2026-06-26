import type { Message } from '@bufbuild/protobuf';
import {
  enumDesc,
  fileDesc,
  messageDesc,
  serviceDesc,
  type GenEnum,
  type GenMessage,
  type GenService,
} from '@bufbuild/protobuf/codegenv2';
import {
  file_google_protobuf_timestamp,
  type Timestamp,
} from '@bufbuild/protobuf/wkt';

// Generated from proto/ecommerce/recommendation/v1/recommendation.proto.
export const file_ecommerce_recommendation_v1_recommendation = fileDesc(
  [
    'CjBlY29tbWVyY2UvcmVjb21tZW5kYXRpb24vdjEvcmVjb21tZW5kYXRpb24ucHJvdG8SG2Vjb21t',
    'ZXJjZS5yZWNvbW1lbmRhdGlvbi52MRofZ29vZ2xlL3Byb3RvYnVmL3RpbWVzdGFtcC5wcm90byKH',
    'AwoZR2V0UmVjb21tZW5kYXRpb25zUmVxdWVzdBIXCgd1c2VyX2lkGAEgASgJUgZ1c2VySWQSTAoH',
    'Y29udGV4dBgCIAEoDjIyLmVjb21tZXJjZS5yZWNvbW1lbmRhdGlvbi52MS5SZWNvbW1lbmRhdGlv',
    'bkNvbnRleHRSB2NvbnRleHQSIQoMYW5vbnltb3VzX2lkGAMgASgJUgthbm9ueW1vdXNJZBIdCgpw',
    'cm9kdWN0X2lkGAQgASgJUglwcm9kdWN0SWQSHwoLY2F0ZWdvcnlfaWQYBSABKAlSCmNhdGVnb3J5',
    'SWQSGwoJc2VsbGVyX2lkGAYgASgJUghzZWxsZXJJZBIoChBjYXJ0X3Byb2R1Y3RfaWRzGAcgAygJ',
    'Ug5jYXJ0UHJvZHVjdElkcxIUCgVsaW1pdBgIIAEoBVIFbGltaXQSQwoEdHlwZRgJIAEoDjIvLmVj',
    'b21tZXJjZS5yZWNvbW1lbmRhdGlvbi52MS5SZWNvbW1lbmRhdGlvblR5cGVSBHR5cGUidQoSUmVj',
    'b21tZW5kYXRpb25JdGVtEh0KCnByb2R1Y3RfaWQYASABKAlSCXByb2R1Y3RJZBISCgRyYW5rGAIg',
    'ASgFUgRyYW5rEhQKBXNjb3JlGAMgASgBUgVzY29yZRIWCgZyZWFzb24YBCABKAlSBnJlYXNvbiLh',
    'AgoaR2V0UmVjb21tZW5kYXRpb25zUmVzcG9uc2USKwoRcmVjb21tZW5kYXRpb25faWQYASABKAlS',
    'EHJlY29tbWVuZGF0aW9uSWQSQwoEdHlwZRgCIAEoDjIvLmVjb21tZXJjZS5yZWNvbW1lbmRhdGlv',
    'bi52MS5SZWNvbW1lbmRhdGlvblR5cGVSBHR5cGUSHwoLc3RyYXRlZ3lfaWQYAyABKAlSCnN0cmF0',
    'ZWd5SWQSRQoFaXRlbXMYBCADKAsyLy5lY29tbWVyY2UucmVjb21tZW5kYXRpb24udjEuUmVjb21t',
    'ZW5kYXRpb25JdGVtUgVpdGVtcxI9CgxnZW5lcmF0ZWRfYXQYBSABKAsyGi5nb29nbGUucHJvdG9i',
    'dWYuVGltZXN0YW1wUgtnZW5lcmF0ZWRBdBIqChFjYWNoZV90dGxfc2Vjb25kcxgGIAEoA1IPY2Fj',
    'aGVUdGxTZWNvbmRzKt8BChJSZWNvbW1lbmRhdGlvblR5cGUSIwofUkVDT01NRU5EQVRJT05fVFlQ',
    'RV9VTlNQRUNJRklFRBAAEigKJFJFQ09NTUVOREFUSU9OX1RZUEVfU0lNSUxBUl9QUk9EVUNUUxAB',
    'EiAKHFJFQ09NTUVOREFUSU9OX1RZUEVfVFJFTkRJTkcQAhIkCiBSRUNPTU1FTkRBVElPTl9UWVBF',
    'X1BFUlNPTkFMSVpFRBADEjIKLlJFQ09NTUVOREFUSU9OX1RZUEVfRlJFUVVFTlRMWV9CT1VHSFRf',
    'VE9HRVRIRVIQBCrXAgoVUmVjb21tZW5kYXRpb25Db250ZXh0EiYKIlJFQ09NTUVOREFUSU9OX0NP',
    'TlRFWFRfVU5TUEVDSUZJRUQQABIkCiBSRUNPTU1FTkRBVElPTl9DT05URVhUX0hPTUVfRkVFRBAB',
    'EikKJVJFQ09NTUVOREFUSU9OX0NPTlRFWFRfUFJPRFVDVF9ERVRBSUwQAhIrCidSRUNPTU1FTkRB',
    'VElPTl9DT05URVhUX0NBVEVHT1JZX0xJU1RJTkcQAxIfChtSRUNPTU1FTkRBVElPTl9DT05URVhU',
    'X0NBUlQQBBIjCh9SRUNPTU1FTkRBVElPTl9DT05URVhUX0NIRUNLT1VUEAUSKQolUkVDT01NRU5E',
    'QVRJT05fQ09OVEVYVF9TRUFSQ0hfUkVTVUxUUxAGEicKI1JFQ09NTUVOREFUSU9OX0NPTlRFWFRf',
    'U0VMTEVSX1NUT1JFEAcynwEKFVJlY29tbWVuZGF0aW9uU2VydmljZRKFAQoSR2V0UmVjb21tZW5k',
    'YXRpb25zEjYuZWNvbW1lcmNlLnJlY29tbWVuZGF0aW9uLnYxLkdldFJlY29tbWVuZGF0aW9uc1Jl',
    'cXVlc3QaNy5lY29tbWVyY2UucmVjb21tZW5kYXRpb24udjEuR2V0UmVjb21tZW5kYXRpb25zUmVz',
    'cG9uc2VCaVpnZ2l0aHViLmNvbS9leGFtcGxlL2Vjb21tZXJjZS1wbGF0Zm9ybS9iYWNrZW5kL3By',
    'b3RvLWdlbi9nby9lY29tbWVyY2UvcmVjb21tZW5kYXRpb24vdjE7cmVjb21tZW5kYXRpb252MWIG',
    'cHJvdG8z',
  ].join(''),
  [file_google_protobuf_timestamp],
);

export enum RecommendationType {
  UNSPECIFIED = 0,
  SIMILAR_PRODUCTS = 1,
  TRENDING = 2,
  PERSONALIZED = 3,
  FREQUENTLY_BOUGHT_TOGETHER = 4,
}

export type RecommendationTypeJson =
  | 'RECOMMENDATION_TYPE_UNSPECIFIED'
  | 'RECOMMENDATION_TYPE_SIMILAR_PRODUCTS'
  | 'RECOMMENDATION_TYPE_TRENDING'
  | 'RECOMMENDATION_TYPE_PERSONALIZED'
  | 'RECOMMENDATION_TYPE_FREQUENTLY_BOUGHT_TOGETHER';

export const RecommendationTypeSchema: GenEnum<
  RecommendationType,
  RecommendationTypeJson
> = enumDesc(file_ecommerce_recommendation_v1_recommendation, 0);

export enum RecommendationContext {
  UNSPECIFIED = 0,
  HOME_FEED = 1,
  PRODUCT_DETAIL = 2,
  CATEGORY_LISTING = 3,
  CART = 4,
  CHECKOUT = 5,
  SEARCH_RESULTS = 6,
  SELLER_STORE = 7,
}

export type RecommendationContextJson =
  | 'RECOMMENDATION_CONTEXT_UNSPECIFIED'
  | 'RECOMMENDATION_CONTEXT_HOME_FEED'
  | 'RECOMMENDATION_CONTEXT_PRODUCT_DETAIL'
  | 'RECOMMENDATION_CONTEXT_CATEGORY_LISTING'
  | 'RECOMMENDATION_CONTEXT_CART'
  | 'RECOMMENDATION_CONTEXT_CHECKOUT'
  | 'RECOMMENDATION_CONTEXT_SEARCH_RESULTS'
  | 'RECOMMENDATION_CONTEXT_SELLER_STORE';

export const RecommendationContextSchema: GenEnum<
  RecommendationContext,
  RecommendationContextJson
> = enumDesc(file_ecommerce_recommendation_v1_recommendation, 1);

export type GetRecommendationsRequest =
  Message<'ecommerce.recommendation.v1.GetRecommendationsRequest'> & {
    userId: string;
    context: RecommendationContext;
    anonymousId: string;
    productId: string;
    categoryId: string;
    sellerId: string;
    cartProductIds: string[];
    limit: number;
    type: RecommendationType;
  };

export type RecommendationItem =
  Message<'ecommerce.recommendation.v1.RecommendationItem'> & {
    productId: string;
    rank: number;
    score: number;
    reason: string;
  };

export type GetRecommendationsResponse =
  Message<'ecommerce.recommendation.v1.GetRecommendationsResponse'> & {
    recommendationId: string;
    type: RecommendationType;
    strategyId: string;
    items: RecommendationItem[];
    generatedAt?: Timestamp | undefined;
    cacheTtlSeconds: bigint;
  };

export const GetRecommendationsRequestSchema: GenMessage<GetRecommendationsRequest> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 0);

export const RecommendationItemSchema: GenMessage<RecommendationItem> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 1);

export const GetRecommendationsResponseSchema: GenMessage<GetRecommendationsResponse> =
  messageDesc(file_ecommerce_recommendation_v1_recommendation, 2);

export const RecommendationService: GenService<{
  getRecommendations: {
    methodKind: 'unary';
    input: typeof GetRecommendationsRequestSchema;
    output: typeof GetRecommendationsResponseSchema;
  };
}> = serviceDesc(file_ecommerce_recommendation_v1_recommendation, 0);
