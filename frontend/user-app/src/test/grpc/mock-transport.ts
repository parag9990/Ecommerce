import {
  create,
  type DescMessage,
  type DescMethodStreaming,
  type DescMethodUnary,
  type MessageInitShape,
} from '@bufbuild/protobuf';
import type {
  ContextValues,
  StreamResponse,
  Transport,
  UnaryResponse,
} from '@connectrpc/connect';
import { Code, ConnectError } from '@connectrpc/connect';

type MockGrpcResponse = MessageInitShape<DescMessage>;
type UnaryHandler = (input: unknown) => MockGrpcResponse | Promise<MockGrpcResponse>;

export type MockGrpcHandlers = Readonly<Record<string, UnaryHandler>>;

export function grpcMethodKey(serviceTypeName: string, methodName: string): string {
  return `${serviceTypeName}/${methodName}`;
}

export function createMockGrpcTransport(handlers: MockGrpcHandlers): Transport {
  return {
    stream<I extends DescMessage, O extends DescMessage>(
      method: DescMethodStreaming<I, O>,
      signal: AbortSignal | undefined,
      timeoutMs: number | undefined,
      header: HeadersInit | undefined,
      input: AsyncIterable<MessageInitShape<I>>,
      contextValues?: ContextValues,
    ): Promise<StreamResponse<I, O>> {
      void method;
      void signal;
      void timeoutMs;
      void header;
      void input;
      void contextValues;

      return Promise.reject(
        new ConnectError(
          'Mock gRPC transport only supports unary calls.',
          Code.Unimplemented,
        ),
      );
    },
    async unary<I extends DescMessage, O extends DescMessage>(
      method: DescMethodUnary<I, O>,
      signal: AbortSignal | undefined,
      timeoutMs: number | undefined,
      header: HeadersInit | undefined,
      input: MessageInitShape<I>,
      contextValues?: ContextValues,
    ): Promise<UnaryResponse<I, O>> {
      void signal;
      void timeoutMs;
      void header;
      void contextValues;

      const handler = handlers[grpcMethodKey(method.parent.typeName, method.name)];

      if (!handler) {
        throw new ConnectError(
          `No mock handler registered for ${method.parent.typeName}/${method.name}.`,
          Code.Unimplemented,
        );
      }

      const output = await handler(input);

      return {
        header: new Headers(),
        message: create(method.output, output as MessageInitShape<O>),
        method,
        service: method.parent,
        stream: false,
        trailer: new Headers(),
      };
    },
  };
}
