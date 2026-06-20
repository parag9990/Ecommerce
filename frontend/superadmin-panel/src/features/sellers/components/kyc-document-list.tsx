import { DataState } from "../../../components/ui/data-state";
import { DocumentPreview } from "../../../components/ui/document-preview";
import { useSellerKycDocuments } from "../hooks/use-seller-kyc-documents";

export function KycDocumentList({
  sellerId,
  canViewKyc,
  canViewMaskedKyc
}: {
  sellerId: string;
  canViewKyc: boolean;
  canViewMaskedKyc: boolean;
}) {
  const documentsQuery = useSellerKycDocuments(sellerId, canViewKyc);

  if (!canViewKyc && canViewMaskedKyc) {
    return (
      <section className="space-y-3">
        <h2 className="text-base font-semibold text-slate-950">KYC Documents</h2>
        <DocumentPreview title="KYC documents" status="masked" masked />
      </section>
    );
  }

  if (!canViewKyc) {
    return null;
  }

  if (documentsQuery.isLoading) {
    return <DataState title="Loading KYC documents" description="Fetching document metadata." />;
  }

  if (documentsQuery.error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load KYC documents"
        description={
          documentsQuery.error instanceof Error
            ? documentsQuery.error.message
            : "The KYC document request failed."
        }
        action={
          <button
            type="button"
            onClick={() => void documentsQuery.refetch()}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  const documents = documentsQuery.data?.documents ?? [];

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-base font-semibold text-slate-950">KYC Documents</h2>
        <p className="mt-1 text-sm text-slate-600">Secure document metadata for admin review.</p>
      </div>

      {documents.length === 0 ? (
        <DataState title="No KYC documents" description="This seller has not submitted review documents yet." />
      ) : (
        <div className="grid gap-3">
          {documents.map((document) => (
            <DocumentPreview
              key={document.document_id}
              title={document.document_type}
              status={document.status}
              url={document.storage_url}
              createdAt={document.created_at}
              reviewedAt={document.reviewed_at}
              rejectionReason={document.rejection_reason}
            />
          ))}
        </div>
      )}
    </section>
  );
}
