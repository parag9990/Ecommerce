import {
  getErrorDetailMessages,
  getRequestId,
  getSafeErrorMessage,
} from "../../lib/api-error";

type FormErrorProps = {
  error: unknown;
};

export function FormError({ error }: FormErrorProps) {
  const requestId = getRequestId(error);
  const details = getErrorDetailMessages(error);

  return (
    <div
      className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-900"
      role="alert"
    >
      <p className="font-medium">{getSafeErrorMessage(error)}</p>
      {details.length ? (
        <ul className="mt-2 list-disc space-y-1 pl-5 text-xs text-red-800">
          {details.map((detail) => (
            <li key={detail}>{detail}</li>
          ))}
        </ul>
      ) : null}
      {requestId ? (
        <p className="mt-1 break-all font-mono text-xs text-red-700">
          Request ID: {requestId}
        </p>
      ) : null}
    </div>
  );
}
