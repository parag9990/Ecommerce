import type { Message } from '@bufbuild/protobuf';
import {
  fileDesc,
  messageDesc,
  serviceDesc,
  type GenMessage,
  type GenService,
} from '@bufbuild/protobuf/codegenv2';

/**
 * Generated-style descriptor for ecommerce/search/v1/search.proto.
 * Replace this file with Buf output when the canonical proto module lands.
 */
export const file_ecommerce_search_v1_search = fileDesc(
  'CiBlY29tbWVyY2Uvc2VhcmNoL3YxL3NlYXJjaC5wcm90bxITZWNvbW1lcmNlLnNlYXJjaC52MSJBChNBdXRvY29tcGxldGVSZXF1ZXN0EhQKBXF1ZXJ5GAEgASgJUgVxdWVyeRIUCgVsaW1pdBgCIAEoBVIFbGltaXQiWAoWQXV0b2NvbXBsZXRlU3VnZ2VzdGlvbhIUCgV2YWx1ZRgBIAEoCVIFdmFsdWUSEgoEdHlwZRgCIAEoCVIEdHlwZRIUCgVsYWJlbBgDIAEoCVIFbGFiZWwiZQoUQXV0b2NvbXBsZXRlUmVzcG9uc2USTQoLc3VnZ2VzdGlvbnMYASADKAsyKy5lY29tbWVyY2Uuc2VhcmNoLnYxLkF1dG9jb21wbGV0ZVN1Z2dlc3Rpb25SC3N1Z2dlc3Rpb25zMnQKDVNlYXJjaFNlcnZpY2USYwoMQXV0b2NvbXBsZXRlEiguZWNvbW1lcmNlLnNlYXJjaC52MS5BdXRvY29tcGxldGVSZXF1ZXN0GikuZWNvbW1lcmNlLnNlYXJjaC52MS5BdXRvY29tcGxldGVSZXNwb25zZWIGcHJvdG8z',
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
  messageDesc(file_ecommerce_search_v1_search, 0);

export const AutocompleteSuggestionSchema: GenMessage<AutocompleteSuggestion> =
  messageDesc(file_ecommerce_search_v1_search, 1);

export const AutocompleteResponseSchema: GenMessage<AutocompleteResponse> =
  messageDesc(file_ecommerce_search_v1_search, 2);

export const SearchService: GenService<{
  autocomplete: {
    methodKind: 'unary';
    input: typeof AutocompleteRequestSchema;
    output: typeof AutocompleteResponseSchema;
  };
}> = serviceDesc(file_ecommerce_search_v1_search, 0);
