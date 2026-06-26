import type { Message } from '@bufbuild/protobuf';
import {
  fileDesc,
  messageDesc,
  serviceDesc,
  type GenMessage,
  type GenService,
} from '@bufbuild/protobuf/codegenv2';
import { file_google_protobuf_struct } from '@bufbuild/protobuf/wkt';

// Generated from proto/ecommerce/search/v1/search.proto.
export const file_ecommerce_search_v1_search = fileDesc(
  [
    'CiBlY29tbWVyY2Uvc2VhcmNoL3YxL3NlYXJjaC5wcm90bxITZWNvbW1lcmNlLnNlYXJjaC52MRoc',
    'Z29vZ2xlL3Byb3RvYnVmL3N0cnVjdC5wcm90byLxAQoNU2VhcmNoUmVxdWVzdBIUCgVxdWVyeRgB',
    'IAEoCVIFcXVlcnkSSQoHZmlsdGVycxgCIAMoCzIvLmVjb21tZXJjZS5zZWFyY2gudjEuU2VhcmNo',
    'UmVxdWVzdC5GaWx0ZXJzRW50cnlSB2ZpbHRlcnMSEgoEc29ydBgDIAEoCVIEc29ydBISCgRwYWdl',
    'GAQgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZRgFIAEoBVIIcGFnZVNpemUaOgoMRmlsdGVyc0VudHJ5',
    'EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAEilAEKDlNlYXJjaFJl',
    'c3BvbnNlEjgKCHByb2R1Y3RzGAEgAygLMhwuZWNvbW1lcmNlLnNlYXJjaC52MS5Qcm9kdWN0Ughw',
    'cm9kdWN0cxIyCgZmYWNldHMYAiADKAsyGi5lY29tbWVyY2Uuc2VhcmNoLnYxLkZhY2V0UgZmYWNl',
    'dHMSFAoFdG90YWwYAyABKANSBXRvdGFsIo0CCgdQcm9kdWN0Eh0KCnByb2R1Y3RfaWQYASABKAlS',
    'CXByb2R1Y3RJZBIbCglzZWxsZXJfaWQYAiABKAlSCHNlbGxlcklkEhQKBXRpdGxlGAMgASgJUgV0',
    'aXRsZRIgCgtkZXNjcmlwdGlvbhgEIAEoCVILZGVzY3JpcHRpb24SFAoFYnJhbmQYBSABKAlSBWJy',
    'YW5kEh8KC2NhdGVnb3J5X2lkGAYgASgJUgpjYXRlZ29yeUlkEhYKBnN0YXR1cxgHIAEoCVIGc3Rh',
    'dHVzEj8KCHZhcmlhbnRzGAggAygLMiMuZWNvbW1lcmNlLnNlYXJjaC52MS5Qcm9kdWN0VmFyaWFu',
    'dFIIdmFyaWFudHMitAEKDlByb2R1Y3RWYXJpYW50EhAKA3NrdRgBIAEoCVIDc2t1EjcKCmF0dHJp',
    'YnV0ZXMYAiABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgphdHRyaWJ1dGVzEjAKBXByaWNl',
    'GAMgASgLMhouZWNvbW1lcmNlLnNlYXJjaC52MS5Nb25leVIFcHJpY2USJQoOc3RvY2tfcXVhbnRp',
    'dHkYBCABKAVSDXN0b2NrUXVhbnRpdHkiOwoFTW9uZXkSFgoGYW1vdW50GAEgASgBUgZhbW91bnQS',
    'GgoIY3VycmVuY3kYAiABKAlSCGN1cnJlbmN5IlYKBUZhY2V0EhQKBWZpZWxkGAEgASgJUgVmaWVs',
    'ZBI3CgZ2YWx1ZXMYAiADKAsyHy5lY29tbWVyY2Uuc2VhcmNoLnYxLkZhY2V0VmFsdWVSBnZhbHVl',
    'cyI4CgpGYWNldFZhbHVlEhQKBXZhbHVlGAEgASgJUgV2YWx1ZRIUCgVjb3VudBgCIAEoA1IFY291',
    'bnQiQQoTQXV0b2NvbXBsZXRlUmVxdWVzdBIUCgVxdWVyeRgBIAEoCVIFcXVlcnkSFAoFbGltaXQY',
    'AiABKAVSBWxpbWl0IlgKFkF1dG9jb21wbGV0ZVN1Z2dlc3Rpb24SFAoFdmFsdWUYASABKAlSBXZh',
    'bHVlEhIKBHR5cGUYAiABKAlSBHR5cGUSFAoFbGFiZWwYAyABKAlSBWxhYmVsImUKFEF1dG9jb21w',
    'bGV0ZVJlc3BvbnNlEk0KC3N1Z2dlc3Rpb25zGAEgAygLMisuZWNvbW1lcmNlLnNlYXJjaC52MS5B',
    'dXRvY29tcGxldGVTdWdnZXN0aW9uUgtzdWdnZXN0aW9ucyJGChRDcmVhdGVTeW5vbnltUmVxdWVz',
    'dBISCgRyb290GAEgASgJUgRyb290EhoKCHN5bm9ueW1zGAIgAygJUghzeW5vbnltcyJGChNMaXN0',
    'U3lub255bXNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaXplGAIgASgFUghw',
    'YWdlU2l6ZSJeCg1TZWFyY2hTeW5vbnltEh0KCnN5bm9ueW1faWQYASABKAlSCXN5bm9ueW1JZBIS',
    'CgRyb290GAIgASgJUgRyb290EhoKCHN5bm9ueW1zGAMgAygJUghzeW5vbnltcyJWChRMaXN0U3lu',
    'b255bXNSZXNwb25zZRI+CghzeW5vbnltcxgBIAMoCzIiLmVjb21tZXJjZS5zZWFyY2gudjEuU2Vh',
    'cmNoU3lub255bVIIc3lub255bXMylAMKDVNlYXJjaFNlcnZpY2USWQoOU2VhcmNoUHJvZHVjdHMS',
    'Ii5lY29tbWVyY2Uuc2VhcmNoLnYxLlNlYXJjaFJlcXVlc3QaIy5lY29tbWVyY2Uuc2VhcmNoLnYx',
    'LlNlYXJjaFJlc3BvbnNlEmMKDEF1dG9jb21wbGV0ZRIoLmVjb21tZXJjZS5zZWFyY2gudjEuQXV0',
    'b2NvbXBsZXRlUmVxdWVzdBopLmVjb21tZXJjZS5zZWFyY2gudjEuQXV0b2NvbXBsZXRlUmVzcG9u',
    'c2USXgoNQ3JlYXRlU3lub255bRIpLmVjb21tZXJjZS5zZWFyY2gudjEuQ3JlYXRlU3lub255bVJl',
    'cXVlc3QaIi5lY29tbWVyY2Uuc2VhcmNoLnYxLlNlYXJjaFN5bm9ueW0SYwoMTGlzdFN5bm9ueW1z',
    'EiguZWNvbW1lcmNlLnNlYXJjaC52MS5MaXN0U3lub255bXNSZXF1ZXN0GikuZWNvbW1lcmNlLnNl',
    'YXJjaC52MS5MaXN0U3lub255bXNSZXNwb25zZUJPWk1naXRodWIuY29tL3BhcmFnL2Vjb21tZXJj',
    'ZS9iYWNrZW5kL3NoYXJlZC9nZW4vZ28vZWNvbW1lcmNlL3NlYXJjaC92MTtzZWFyY2h2MWIGcHJv',
    'dG8z',
  ].join(''),
  [file_google_protobuf_struct],
);

export type AutocompleteRequest =
  Message<'ecommerce.search.v1.AutocompleteRequest'> & {
    query: string;
    limit: number;
  };

export type AutocompleteSuggestion =
  Message<'ecommerce.search.v1.AutocompleteSuggestion'> & {
    value: string;
    type: string;
    label: string;
  };

export type AutocompleteResponse =
  Message<'ecommerce.search.v1.AutocompleteResponse'> & {
    suggestions: AutocompleteSuggestion[];
  };

export const AutocompleteRequestSchema: GenMessage<AutocompleteRequest> =
  messageDesc(file_ecommerce_search_v1_search, 7);

export const AutocompleteSuggestionSchema: GenMessage<AutocompleteSuggestion> =
  messageDesc(file_ecommerce_search_v1_search, 8);

export const AutocompleteResponseSchema: GenMessage<AutocompleteResponse> =
  messageDesc(file_ecommerce_search_v1_search, 9);

export const SearchService: GenService<{
  autocomplete: {
    methodKind: 'unary';
    input: typeof AutocompleteRequestSchema;
    output: typeof AutocompleteResponseSchema;
  };
}> = serviceDesc(file_ecommerce_search_v1_search, 0);
