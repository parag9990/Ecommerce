import { getRequestId, getSafeErrorMessage } from "../../lib/api-error";

type FormErrorProps = {
  error: unknown;
};

export function FormError({ error }: FormErrorProps) {
  const requestId = getRequestId(error);

  return (
    <div
      className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-900"
      role="alert"
    >
      <p className="font-medium">{getSafeErrorMessage(error)}</p>
      {requestId ? (
        <p className="mt-1 break-all font-mono text-xs text-red-700">
          Request ID: {requestId}
        </p>
      ) : null}
    </div>
  );
}
